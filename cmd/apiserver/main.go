package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dehimb/cake/internal/server"
	"github.com/dehimb/cake/internal/store"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Catch interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-c
		logger.Infof("Signal: %s", sig)
		cancel()
	}()

	/* go func() {
	 *   time.Sleep(5 * time.Second)
	 *   syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	 * }() */

	server.Start(ctx, store.New(ctx, logger, "cake.db"), logger)
}
