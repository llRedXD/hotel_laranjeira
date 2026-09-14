package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hotel-laranjeira/backend/internal/auth"
	"github.com/hotel-laranjeira/backend/internal/config"
	"github.com/hotel-laranjeira/backend/internal/database"
	"github.com/hotel-laranjeira/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	if !cfg.IsDevelopment() && cfg.JWTSecret == "hotel-laranjeira-development-secret-change-me" {
		log.Fatal("JWT_SECRET deve ser configurado fora do ambiente de desenvolvimento")
	}
	ctx := context.Background()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("não foi possível conectar ao banco de dados: %v", err)
	}
	defer db.Close()
	if err := auth.BootstrapUsers(ctx, db, cfg); err != nil {
		log.Fatalf("não foi possível configurar os usuários iniciais: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(db, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("API iniciada na porta %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor HTTP falhou: %v", err)
		}
	}()

	signalContext, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-signalContext.Done()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("erro ao encerrar servidor: %v", err)
	}
}
