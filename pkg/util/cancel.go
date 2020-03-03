package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/develar/app-builder/pkg/log"
	"go.uber.org/zap"
)

func CreateContext() (context.Context, context.CancelFunc) {
	c, cancel := context.WithCancel(context.Background())
	go onCancelSignal(cancel)
	return c, cancel
}

func CreateContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	c, cancel := context.WithTimeout(context.Background(), timeout)
	go onCancelSignal(cancel)
	return c, cancel
}

func onCancelSignal(cancel context.CancelFunc) {
	defer cancel()
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.Info("canceling", zap.String("signal", sig.String()))
}
