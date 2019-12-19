package server

import (
	"context"
	"net/http"
	"time"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func Start(ctx context.Context, store *store.Store, logger *logrus.Logger) {
	handler := &handler{
		router: mux.NewRouter(),
		store:  store,
		logger: logger,
	}
	s := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	handler.initRouter()

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			logger.Fatal("Error starting server", err)
		}
	}()

	logger.Info("Server started")

	waitForShutdown(ctx, s, logger)
	// Wait untill all services can stop
	time.Sleep(1 * time.Second)
	logger.Info("Exiting...")
}

func waitForShutdown(ctx context.Context, s *http.Server, logger *logrus.Logger) {
	<-ctx.Done()
	logger.Info("Trying graceful shutdown server")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctxShutDown); err != http.ErrServerClosed {
		logger.Fatalf("Server shutdown failed: %s", err)
		return
	}
	logger.Info("Server stopped")
}
