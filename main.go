package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func fatal(msg string, err error) {
	panic(fmt.Sprintf("%s, %v", msg, err))
}

func waitForInterrupt() {
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func main() {
	config, err := ReadConfig("recon.yaml")
	if err != nil {
		fatal("couldn't read config", err)
	}

	clientset, err := MakeK8Client()
	if err != nil {
		fatal("failed to create k8 client", err)
	}

	ctx, cancelFn := context.WithCancel(context.Background())

	runner := MakeRunner(&config, clientset.CoreV1().Pods("default"))
	runner.RunLogs(ctx)

	waitForInterrupt()
	cancelFn()
}
