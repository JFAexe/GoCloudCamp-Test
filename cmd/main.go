package main

import (
	"context"
	"fmt"
	"os"

	"gocloudcamp_test/internal/database"
	"gocloudcamp_test/internal/handlers"
	"gocloudcamp_test/internal/server"
	"gocloudcamp_test/internal/service"
)

func main() {
	serviceCtx, cancel := context.WithCancel(context.Background())

	uri := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	database := database.Connect(serviceCtx, uri)
	service := service.New(database)
	handlers := handlers.Get(serviceCtx, service)
	server := server.New(handlers, "0.0.0.0:8080")

	service.Start()

	go server.Run()
	go func() {
		defer cancel()
		server.GracefulShutdown(serviceCtx, service.ChanForceStop)
	}()

	go service.ForceStop(cancel)

	service.Stop(serviceCtx)
}
