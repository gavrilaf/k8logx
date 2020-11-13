package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type StreamerConfig struct {
	K8Provider    v1.PodInterface
	PodName       string
	ContainerName string
	Receiver      *Receiver
}

type Streamer struct {
	k8provider    v1.PodInterface
	podName       string
	containerName string
	receiver      *Receiver
	closed        chan struct{}
}

func MakeStreamer(cfg StreamerConfig) *Streamer {
	return &Streamer{
		k8provider:    cfg.K8Provider,
		podName:       cfg.PodName,
		containerName: cfg.ContainerName,
		receiver:      cfg.Receiver,
		closed:        make(chan struct{}),
	}
}

func (s *Streamer) Run(ctx context.Context) error {
	logOptions := corev1.PodLogOptions{
		Container:                    s.containerName,
		Follow:                       true,
		Previous:                     false,
		SinceSeconds:                 nil,
		SinceTime:                    nil,
		Timestamps:                   false,
		TailLines:                    nil,
		LimitBytes:                   nil,
		InsecureSkipTLSVerifyBackend: false,
	}

	logsReq := s.k8provider.GetLogs(s.podName, &logOptions)

	stream, err := logsReq.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to read log stream, pod: %s, container: %s, %w", s.podName, s.containerName, err)
	}

	go func() {
		<-s.closed
		if err := stream.Close(); err != nil {
			fmt.Printf("failed to close stream, pod: %s, container: %s, %w", s.podName, s.containerName, err)
		}
	}()

	go func() {
		reader := bufio.NewReader(stream)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			line = bytes.TrimSpace(line)

			if len(line) == 0 {
				continue
			}

			s.receiver.Receive(line)
		}
	}()

	go func() {
		<-ctx.Done()
		close(s.closed)
	}()

	return nil
}

func (s *Streamer) Close() {
	s.receiver.Receive([]byte("--------------------"))
	close(s.closed)
}