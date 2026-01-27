package handler

import (
	"net/http"

	"github.com/luispfcanales/apirain/pkg/config"
	mantenimientoHandler "github.com/luispfcanales/apirain/pkg/handler"
	"github.com/luispfcanales/apirain/pkg/response"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig()
	if err != nil {
		response.InternalServerError(w, "Error de configuración")
		return
	}

	h, err := mantenimientoHandler.NewMaintenanceHandler(cfg)
	if err != nil {
		response.InternalServerError(w, "Error inicializando handler")
		return
	}

	h.ListRequests(w, r)
}
