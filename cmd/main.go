package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gocloudcamp_test/internal/database"
	"gocloudcamp_test/internal/handlers"
	"gocloudcamp_test/internal/server"
	"gocloudcamp_test/internal/service"
)

func main() {
	chanQuit := make(chan os.Signal, 1)
	signal.Notify(chanQuit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	serviceCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uri := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	database := database.Connect(uri)
	service := service.New(database)
	handlers := handlers.Get(serviceCtx, service)
	server := server.New(serviceCtx, handlers, "0.0.0.0:8080")

	service.Start()

	go server.Run()
	go server.GracefulShutdown(serviceCtx, cancel, chanQuit)

	service.Stop(serviceCtx)
}
