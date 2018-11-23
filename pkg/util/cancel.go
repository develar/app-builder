package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/apex/log"
)

func CreateContext() (context.Context, context.CancelFunc) {
	downloadContext, cancel := context.WithCancel(context.Background())
	go onCancelSignal(cancel)
	return downloadContext, cancel
}

func onCancelSignal(cancel context.CancelFunc) {
	defer cancel()
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.Infof("%v: canceling...\n", sig)
}
