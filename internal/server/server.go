package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	http.Server
}

func New(addr string, handlers http.Handler) *Server {
	server := &Server{}

	server.Addr = addr
	server.Handler = handlers

	return server
}

func (s *Server) Run() {
	log.Printf("http | server starting | %s", s.Addr)

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("http | %v", err)

		return
	}

	log.Printf("http | shut down")
}

func (s *Server) GracefulShutdown(ctx context.Context, forceStop chan<- struct{}) {
	chanQuit := make(chan os.Signal, 1)
	signal.Notify(chanQuit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-chanQuit

	log.Print("http | shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*5)
	defer shutdownCancel()

	go func() {
		<-shutdownCtx.Done()

		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Print("http | force stop")

			forceStop <- struct{}{}
		}
	}()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("http | %v", err)
	}
}
