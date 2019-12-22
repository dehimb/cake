package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/dehimb/cake/internal/server"
	"github.com/dehimb/cake/internal/store"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Catch interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signal := <-c
		logger.Infof("Signal: %s", signal)
		cancel()
	}()

	server.Start(ctx, store.New(ctx, logger, "cake.db"), logger)
}
