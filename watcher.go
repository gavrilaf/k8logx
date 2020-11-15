package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Target struct {
	podName       string
	containerName string
	containerCfg  *ContainerConfig
}

func (t Target) ID() string {
	return t.podName + ":" + t.containerName
}

type Watcher struct {
	config  *Config
	added   chan Target
	removed chan Target
	closed chan struct{}
}

func MakeWatcher(config *Config) *Watcher {
	return &Watcher{
		config:  config,
		added:   make(chan Target),
		removed: make(chan Target),
		closed:  make(chan struct{}),
	}
}

func (w *Watcher) Added() chan Target {
	return w.added
}

func (w *Watcher) Removed() chan Target {
	return w.removed
}

func (w *Watcher) Run(ctx context.Context, k8 v1.PodInterface) error {
	watcher, err := k8.Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		return fmt.Errorf("failed to create watcher, %w", err)
	}

	go func() {
		fmt.Println("*** watcher started")
		for {
			select {
			case e := <-watcher.ResultChan():
				if e.Object == nil {
					fmt.Println("*** some error in watcher")
					return
				}

				pod := e.Object.(*corev1.Pod)

				switch e.Type {
				case watch.Added, watch.Modified:
					w.resourceAdded(pod)
				case watch.Deleted:
					w.resourceDeleted(pod)
				}
			case <-w.closed:
				watcher.Stop()
				close(w.added)
				close(w.removed)
				fmt.Println("*** watcher stopped")
				return
			}
		}
	}()

	return nil
}

func (w *Watcher) Stop() {
	close(w.closed)
}

func (w *Watcher) resourceAdded(p *corev1.Pod) {
	podCfg := w.getPodConfig(p.Name)
	if podCfg == nil {
		return
	}

	var statuses []corev1.ContainerStatus
	statuses = append(statuses, p.Status.ContainerStatuses...)
	statuses = append(statuses, p.Status.InitContainerStatuses...)

	for _, c := range statuses {
		containerCfg := podCfg.GetContainerConfig(c.Name)
		if containerCfg == nil {
			continue
		}

		target := Target{podName: p.Name, containerName: c.Name, containerCfg: containerCfg}
		w.added <- target
	}
}

func (w *Watcher) resourceDeleted(p *corev1.Pod) {
	if w.getPodConfig(p.Name) == nil {
		return
	}

	var statuses []corev1.ContainerStatus
	statuses = append(statuses, p.Status.ContainerStatuses...)
	statuses = append(statuses, p.Status.InitContainerStatuses...)

	for _, c := range statuses {
		target := Target{podName: p.Name, containerName: c.Name}
		w.removed <- target
	}
}

func (w *Watcher) getPodConfig(name string) *PodConfig {
	for _, p := range w.config.Pods {
		if p.IsPodMatched(name) {
			return &p
		}
	}
	return nil
}
