package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

func requestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()

			defer func() {
				log.Printf(
					"http | request | status %d | method %s | uri %s | address %s | latency %dns | bytes %d",
					ww.Status(),
					r.Method,
					r.RequestURI,
					r.RemoteAddr,
					time.Since(t).Nanoseconds(),
					ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
