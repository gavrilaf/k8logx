package main

import (
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
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func main() {
	cfg, err := ReadConfig("recon.yaml")
	if err != nil {
		fatal("coudn't read config", err)
	}

	fmt.Println(cfg)

	/*clientset, err := MakeK8Client()
	if err != nil {
		fatal("failed to create k8 client", err)
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

	streamer.Close()*/
}

