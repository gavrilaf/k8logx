package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"runtime"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type StreamerConfig struct {
	K8Provider    v1.PodInterface
	PodName       string
	ContainerName string
	Receiver      *Receiver
	Seconds       int
}

type Streamer struct {
	k8provider    v1.PodInterface
	podName       string
	containerName string
	receiver      *Receiver
	closed        chan struct{}
	seconds       int
}

func MakeStreamer(cfg StreamerConfig) *Streamer {
	return &Streamer{
		k8provider:    cfg.K8Provider,
		podName:       cfg.PodName,
		containerName: cfg.ContainerName,
		receiver:      cfg.Receiver,
		closed:        make(chan struct{}),
		seconds:       cfg.Seconds,
	}
}

func (s *Streamer) Run(ctx context.Context) error {
	seconds := int64(s.seconds)
	if seconds == 0 {
		seconds = 300
	}

	logOptions := corev1.PodLogOptions{
		Container:    s.containerName,
		Follow:       true,
		SinceSeconds: &seconds,
	}

	logsReq := s.k8provider.GetLogs(s.podName, &logOptions)

	stream, err := logsReq.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to read log stream, pod: %s, container: %s, %w", s.podName, s.containerName, err)
	}

	go func() {
		<-s.closed
		if err := stream.Close(); err != nil {
			fmt.Printf("failed to close stream, pod: %s, container: %s, %v", s.podName, s.containerName, err)
		}
	}()

	go func() {
		reader := bufio.NewReader(stream)
		s.receiver.printLine([]byte("connected"))
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

			runtime.Gosched()
		}
	}()

	return nil
}

func (s *Streamer) Close() {
	s.receiver.Close()
	close(s.closed)
}
