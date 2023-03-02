package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	http.Server
}

func New(ctx context.Context, handlers http.Handler, addr string) *Server {
	server := &Server{}

	server.Addr = addr
	server.Handler = handlers

	return server
}

func (s *Server) Run() {
	log.Printf("http | server starting | %s", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		switch err {
		case http.ErrServerClosed:
			log.Print("http | shut down")
		default:
			log.Fatalf("http | %v", err)
		}
	}
}

func (s *Server) GracefulShutdown(ctx context.Context, cancel context.CancelFunc, quit <-chan os.Signal) {
	<-quit

	log.Print("http | shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*5)
	defer shutdownCancel()

	go func() {
		<-shutdownCtx.Done()

		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("http | force exit")
		}
	}()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("http | %v", err)
	}

	cancel()
}
