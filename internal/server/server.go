package server

import (
	"context"
	"net/http"
	"time"

	"github.com/dehimb/cake/internal/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func Start(ctx context.Context, storeHandler store.StoreHandler, logger *logrus.Logger) {
	handler := &handler{
		router:       mux.NewRouter(),
		storeHandler: storeHandler,
		logger:       logger,
	}
	s := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	m := &middleware{
		logger:       logger,
		storeHandler: storeHandler,
	}
	handler.initRouter(m)

	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting server", err)
		}
	}()

	logger.Info("Server started")

	waitForShutdown(ctx, s, logger)
	// Wait untill all services can stop
	// TODO change this logic
	time.Sleep(1 * time.Second)
	logger.Info("Exiting...")
}

func waitForShutdown(ctx context.Context, s *http.Server, logger *logrus.Logger) {
	<-ctx.Done()
	logger.Info("Trying graceful shutdown server")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctxShutDown); err != nil {
		logger.Errorf("Server shutdown failed: %s", err)
		return
	}
	logger.Info("Server stopped")
}
