package main

import (
	"context"
	"fmt"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	"os/signal"
	"syscall"
)

func fatal(msg string, err error) {
	panic(fmt.Sprintf("%s, %v", msg, err))
}

func waitForInterrupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func main() {
	clientset, err := MakeK8Client()
	if err != nil {
		panic(err.Error())
	}
	
	ctx := context.Background()
	
	receiver := &Receiver{}
	
	cfg := StreamerConfig{
		K8Provider:    clientset.CoreV1().Pods("default"),
		PodName:       "api-deployment-76fd458bcd-pkt9t",
		ContainerName: "recon-api-app",
		Receiver:      receiver,
	}
	streamer := MakeStreamer(cfg)

	streamer.Run(ctx)

	waitForInterrupt()

	streamer.Close()
}

