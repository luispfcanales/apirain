package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/luispfcanales/apirain/pkg/config"
	"github.com/luispfcanales/apirain/pkg/data"
	"github.com/luispfcanales/apirain/pkg/models"
	"github.com/luispfcanales/apirain/pkg/response"
)

type MaintenanceHandler struct {
	odoo    *data.OdooClient
	odooDev *data.OdooClient
}

type GroupedRequestsWithTeam struct {
	Team     models.MaintenanceTeam      `json:"team"`
	Requests []models.MaintenanceRequest `json:"requests"`
	Total    int                         `json:"total"`
}

func NewMaintenanceHandler(cfg *config.Config) (*MaintenanceHandler, error) {
	odoo, err := data.NewOdooClient(cfg.OdooURL, cfg.OdooDB, cfg.OdooUsername, cfg.OdooPassword)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Odoo Prod: %w", err)
	}

	odooDev, err := data.NewOdooClient(cfg.OdooURLDev, cfg.OdooDBDev, cfg.OdooUsername, cfg.OdooPasswordDev)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Odoo Dev: %w", err)
	}

	return &MaintenanceHandler{
		odoo:    odoo,
		odooDev: odooDev,
	}, nil
}

func (h *MaintenanceHandler) getOdooClient(r *http.Request) *data.OdooClient {
	base := r.URL.Query().Get("base")
	if base == "prod" {
		return h.odoo
	}
	return h.odooDev
}

func (h *MaintenanceHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	response.SetupCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	client := h.getOdooClient(r)
	teams, err := client.GetMaintenanceTeams()
	if err != nil {
		log.Printf("Error [ListTeams]: %v", err)
		response.InternalServerError(w, "Error obteniendo equipos")
		return
	}
	response.JSON(w, http.StatusOK, teams)
}

// func (h *MaintenanceHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
// 	requests, err := h.odoo.GetMaintenanceRequests()
// 	if err != nil {
// 		response.InternalServerError(w, "Error obteniendo solicitudes")
// 		return
// 	}

// 	// Conversión de zona horaria (UTC a PET -5)
// 	loc, err := time.LoadLocation("America/Lima")
// 	if err != nil {
// 		// Fallback si no se puede cargar la ubicación
// 		loc = time.FixedZone("PET", -5*60*60)
// 	}

// 	for i := range requests {
// 		if dateStr, ok := requests[i].ScheduleDate.(string); ok && dateStr != "" {
// 			// El formato de Odoo suele ser "2006-01-02 15:04:05" en UTC
// 			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.UTC)
// 			if err == nil {
// 				requests[i].ScheduleDate = t.In(loc).Format("2006-01-02 15:04:05")
// 			}
// 		}
// 		if dateStr, ok := requests[i].CorrectiveDate.(string); ok && dateStr != "" {
// 			// El formato de Odoo suele ser "2006-01-02 15:04:05" en UTC
// 			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.UTC)
// 			if err == nil {
// 				requests[i].CorrectiveDate = t.In(loc).Format("2006-01-02 15:04:05")
// 			}
// 		}
// 	}

//		response.JSON(w, http.StatusOK, requests)
//	}

// toFloat64 convierte un valor any a float64 de forma segura.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// toString extrae un string de un campo any (puede ser string o false en Odoo).
func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// clamp100 limita un valor al rango [0, 100] y redondea a 2 decimales.
func clamp100(v float64) float64 {
	if v > 100 {
		v = 100
	}
	if v < 0 {
		v = 0
	}
	return round2(v)
}

// round2 redondea un float64 a 2 decimales.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// calculateProgress calcula el porcentaje de avance de una solicitud de mantenimiento.
// Las fechas de ScheduleDate y CorrectiveDate ya deben estar convertidas a hora local (loc).
func calculateProgress(req *models.MaintenanceRequest, loc *time.Location) float64 {
	recurrenceType := toString(req.RecurrenceType)
	repeatUnit := toString(req.RepeatUnit)

	// --- POR HORAS ---
	if recurrenceType == "hours" || recurrenceType == "hour" ||
		repeatUnit == "hours" || repeatUnit == "hour" {
		usedValue, okUsed := toFloat64(req.UsedValue)
		recurrenceValue, okRec := toFloat64(req.RecurrenceValue)
		if !okUsed || !okRec || recurrenceValue == 0 {
			return 0
		}
		return clamp100((usedValue / recurrenceValue) * 100)
	}

	// --- POR FECHA ---
	const dateFmt = "2006-01-02 15:04:05"
	const dayFmt = "2006-01-02"

	// Fecha inicio: request_date (solo fecha) o corrective_date
	var startDate time.Time
	hasStart := false
	if rdStr := toString(req.RequestDate); rdStr != "" {
		t, err := time.ParseInLocation(dayFmt, rdStr, loc)
		if err == nil {
			startDate = t
			hasStart = true
		}
	}
	if !hasStart {
		if cdStr := toString(req.CorrectiveDate); cdStr != "" {
			t, err := time.ParseInLocation(dateFmt, cdStr, loc)
			if err == nil {
				startDate = t
				hasStart = true
			}
		}
	}

	// Fecha fin: preventive_date, sino schedule_date, sino corrective_date
	var targetDate time.Time
	hasTarget := false
	if pdStr := toString(req.PreventiveDate); pdStr != "" {
		t, err := time.ParseInLocation(dateFmt, pdStr, loc)
		if err == nil {
			targetDate = t
			hasTarget = true
		}
	}
	if !hasTarget {
		if sdStr, ok := req.ScheduleDate.(string); ok && sdStr != "" {
			t, err := time.ParseInLocation(dateFmt, sdStr, loc)
			if err == nil {
				targetDate = t
				hasTarget = true
			}
		}
	}
	if !hasTarget {
		if cdStr := toString(req.CorrectiveDate); cdStr != "" {
			t, err := time.ParseInLocation(dateFmt, cdStr, loc)
			if err == nil {
				targetDate = t
				hasTarget = true
			}
		}
	}

	if !hasStart || !hasTarget {
		return 0
	}

	now := time.Now().In(loc)

	if now.After(targetDate) || now.Equal(targetDate) {
		return 100
	}

	totalDuration := targetDate.Sub(startDate)
	if totalDuration <= 0 {
		return 0
	}

	elapsed := now.Sub(startDate)
	if elapsed < 0 {
		elapsed = 0
	}

	return clamp100((float64(elapsed) / float64(totalDuration)) * 100)
}

func (h *MaintenanceHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	response.SetupCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	client := h.getOdooClient(r)

	// Obtener todos los equipos de mantenimiento
	teams, err := client.GetMaintenanceTeams()
	if err != nil {
		log.Printf("Error [ListRequests - GetMaintenanceTeams]: %v", err)
		response.InternalServerError(w, "Error obteniendo equipos de mantenimiento")
		return
	}

	// Obtener todas las solicitudes
	requests, err := client.GetMaintenanceRequests()
	if err != nil {
		log.Printf("Error [ListRequests - GetMaintenanceRequests]: %v", err)
		response.InternalServerError(w, "Error obteniendo solicitudes")
		return
	}

	// Conversión de zona horaria (UTC a PET -5)
	loc, err := time.LoadLocation("America/Lima")
	if err != nil {
		loc = time.FixedZone("PET", -5*60*60)
	}

	for i := range requests {
		if dateStr, ok := requests[i].ScheduleDate.(string); ok && dateStr != "" {
			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.UTC)
			if err == nil {
				requests[i].ScheduleDate = t.In(loc).Format("2006-01-02 15:04:05")
			}
		}
		if dateStr, ok := requests[i].CorrectiveDate.(string); ok && dateStr != "" {
			t, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.UTC)
			if err == nil {
				requests[i].CorrectiveDate = t.In(loc).Format("2006-01-02 15:04:05")
			}
		}

		// Redondear y formatear el valor de uso (como string para mantener 33.30)
		if usedVal, ok := toFloat64(requests[i].UsedValue); ok {
			requests[i].UsedValue = fmt.Sprintf("%.2f", usedVal)
		}

		// Calcular progreso DESPUÉS de la conversión de zona horaria
		requests[i].Progress = calculateProgress(&requests[i], loc)
	}

	// Agrupar las solicitudes por equipo
	groupedRequests := GroupRequestsByTeamWithTeamObject(teams, requests)

	response.JSON(w, http.StatusOK, groupedRequests)
}

// GroupRequestsByTeamWithTeamObject agrupa las solicitudes usando el objeto Team completo
func GroupRequestsByTeamWithTeamObject(teams []models.MaintenanceTeam, requests []models.MaintenanceRequest) []GroupedRequestsWithTeam {
	// Crear un mapa para organizar las solicitudes por team_id
	requestsByTeam := make(map[int][]models.MaintenanceRequest)

	// Organizar las solicitudes en el mapa
	for _, req := range requests {
		teamID := req.GetTeamID()
		requestsByTeam[teamID] = append(requestsByTeam[teamID], req)
	}

	// Crear la estructura agrupada basada en los equipos
	result := make([]GroupedRequestsWithTeam, 0, len(teams))

	// Procesar cada equipo del sistema
	for _, team := range teams {
		teamRequests := requestsByTeam[team.ID]

		group := GroupedRequestsWithTeam{
			Team:     team,
			Requests: teamRequests,
			Total:    len(teamRequests),
		}

		result = append(result, group)
	}

	// Agregar un grupo para solicitudes sin equipo asignado (si las hay)
	if unassignedRequests, exists := requestsByTeam[0]; exists && len(unassignedRequests) > 0 {
		unassignedTeam := models.MaintenanceTeam{
			ID:   0,
			Name: "Sin equipo asignado",
		}

		unassignedGroup := GroupedRequestsWithTeam{
			Team:     unassignedTeam,
			Requests: unassignedRequests,
			Total:    len(unassignedRequests),
		}

		result = append(result, unassignedGroup)
	}

	// Ordenar por total de solicitudes (descendente)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Total > result[j].Total
	})

	return result
}

func (h *MaintenanceHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	response.SetupCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	client := h.getOdooClient(r)
	equipment, err := client.GetMaintenanceEquipment()
	if err != nil {
		log.Printf("Error [ListEquipment]: %v", err)
		response.InternalServerError(w, "Error obteniendo activos")
		return
	}
	response.JSON(w, http.StatusOK, equipment)
}
