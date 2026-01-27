package handler

import (
	"net/http"
	"sort"
	"time"

	"github.com/luispfcanales/apirain/pkg/config"
	"github.com/luispfcanales/apirain/pkg/data"
	"github.com/luispfcanales/apirain/pkg/models"
	"github.com/luispfcanales/apirain/pkg/response"
)

type MaintenanceHandler struct {
	odoo *data.OdooClient
}

type GroupedRequestsWithTeam struct {
	Team     models.MaintenanceTeam      `json:"team"`
	Requests []models.MaintenanceRequest `json:"requests"`
	Total    int                         `json:"total"`
}

func NewMaintenanceHandler(cfg *config.Config) (*MaintenanceHandler, error) {
	odoo, err := data.NewOdooClient(cfg)
	if err != nil {
		return nil, err
	}
	return &MaintenanceHandler{odoo: odoo}, nil
}

func (h *MaintenanceHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	response.SetupCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	teams, err := h.odoo.GetMaintenanceTeams()
	if err != nil {
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
func (h *MaintenanceHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	response.SetupCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Obtener todos los equipos de mantenimiento
	teams, err := h.odoo.GetMaintenanceTeams()
	if err != nil {
		response.InternalServerError(w, "Error obteniendo equipos de mantenimiento")
		return
	}

	// Obtener todas las solicitudes
	requests, err := h.odoo.GetMaintenanceRequests()
	if err != nil {
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

	equipment, err := h.odoo.GetMaintenanceEquipment()
	if err != nil {
		response.InternalServerError(w, "Error obteniendo activos")
		return
	}
	response.JSON(w, http.StatusOK, equipment)
}
