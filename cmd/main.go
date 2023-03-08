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
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	addr := fmt.Sprintf(
		"0.0.0.0:%s",
		os.Getenv("SERVICE_PORT"),
	)

	database := database.Connect(serviceCtx, uri)
	service := service.New(database)
	handlers := handlers.New(serviceCtx, service)
	server := server.New(addr, handlers)

	service.Start()

	go server.Run()
	go func() {
		defer cancel()
		server.GracefulShutdown(serviceCtx, service.ChanForceStop)
	}()

	go service.ForceStop(cancel)

	service.Stop(serviceCtx)
}
