# k8logx

Kubernetes logs viewer supports streaming from several containers and json logs parsing.

- parses json logs and outputs its in more human friendly format
- supports streaming logs from several containers
- monitors kubernetes and dynamically adds and removes containers

## How to run
git clone https://github.com/gavrilaf/k8logx.git
go build
./k8logx

If runned without params utility starts stream logs for the all pods and all containers in the current context.

## Config

Supports config files in yaml format:

```
---
namespace: "default"
seconds-before: 600
pods:
  - api:
    pattern: "api-deployment"
    containers:
      - api:
        pattern: "api-app"
        fields-order:
          - ["method", "uri", "status", "latency"]
          - ["sql"]
  - dispatcher:
    pattern: "dispatcher-deployment"
    containers:
      - dispatcher:
        pattern: "dispatcher"
```

`namespace` - kube namespace for monitoring
`seconds-before` - since seconds from now
`pods` - you can define pods you want to monitor (if section is empty k8logx will monitor all available pods)
