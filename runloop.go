package main

import (
	"context"
	"sync"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type logPair struct {
	streamer *Streamer
	receiver *Receiver
}

type Runner struct {
	k8      v1.PodInterface
	config  *Config
	watcher *Watcher
	streams map[string]logPair
	lock    *sync.Mutex
	added   chan Target
	removed chan Target
}

func MakeRunner(config *Config, k8 v1.PodInterface) *Runner {
	added := make(chan Target)
	removed := make(chan Target)

	watcher := &Watcher{
		config:  config,
		added:   added,
		removed: removed,
	}

	return &Runner{
		k8:      k8,
		config:  config,
		watcher: watcher,
		streams: map[string]logPair{},
		lock:    &sync.Mutex{},
		added:   added,
		removed: removed,
	}
}

func (r *Runner) RunLogs(ctx context.Context) error {
	go func() {
		for target := range r.watcher.Added() {
			r.addTarget(target)
		}
	}()

	go func() {
		for target := range r.watcher.Removed() {
			r.removeTarget(target)
		}
	}()

	return r.watcher.Run(ctx, r.k8)
}

func (r *Runner) addTarget(target Target) {
	r.lock.Lock()
	defer r.lock.Unlock()

	parser := MakeParser(r.config.Mapping)
	receiver := MakeReceiver(target.podName, target.containerName, len(r.streams), target.containerCfg.Fields, parser)

	streamer := MakeStreamer(StreamerConfig{
		K8Provider:    r.k8,
		PodName:       target.podName,
		ContainerName: target.containerName,
		Receiver:      receiver,
		Seconds:       r.config.SecondsBefore,
	})

	if err := streamer.Run(context.Background()); err != nil {
		//fmt.Printf("failed to run streamer for %s, %v\n", target.ID(), err)
		return
	}
	r.streams[target.ID()] = logPair{streamer: streamer, receiver: receiver}
}

func (r *Runner) removeTarget(target Target) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pair, ok := r.streams[target.ID()]
	if !ok {
		return
	}

	pair.streamer.Close()
	delete(r.streams, target.ID())
}

func (r *Runner) Stop() {
	r.watcher.Stop()
}
