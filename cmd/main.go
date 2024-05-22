package main

import (
	"context"
	"github.com/binsabit/resti_tz/config"
	"github.com/binsabit/resti_tz/internal/account"
	"github.com/binsabit/resti_tz/internal/api"
	"github.com/binsabit/resti_tz/internal/http"
	database "github.com/binsabit/resti_tz/internal/repository"
	"github.com/binsabit/resti_tz/internal/transactions"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad("./config.yaml")

	db := database.NewDatabase(context.TODO(), cfg.Database)

	handler := api.NewHandler(db, transactions.Repository{}, account.Repository{})

	server := http.NewServer(cfg.Http, handler)

	go func() {
		log.Fatal(server.Start())
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
}
