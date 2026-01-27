package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luispfcanales/apirain/pkg/config"
	"github.com/luispfcanales/apirain/pkg/handler"
)

func main() {
	// 1. Cargar Configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error de configuración: %v", err)
	}

	// 2. Inicializar Handler principal que contiene la lógica
	h, err := handler.NewMaintenanceHandler(cfg)
	if err != nil {
		log.Fatalf("Error al conectar con Odoo: %v", err)
	}
	fmt.Println("✓ Conexión con Odoo exitosa")

	// 3. Registrar rutas locales
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/mantenimiento/teams", h.ListTeams)
	mux.HandleFunc("GET /api/mantenimiento/requests", h.ListRequests)
	mux.HandleFunc("GET /api/mantenimiento/equipment", h.ListEquipment)

	// 4. Configurar Servidor
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Manejar Apagado Smooth (Graceful Shutdown)
	go func() {
		fmt.Printf("Servidor local escuchando en http://localhost:%s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en el servidor: %v", err)
		}
	}()

	// Esperar señal de interrupción
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nApagando el servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error durante el apagado: %v", err)
	}

	fmt.Println("Servidor detenido correctamente")
}
