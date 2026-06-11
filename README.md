# hs-mcp-manager-core-backend

A Go-based orchestration backend designed to manage, edit, and test multiple Model Context Protocol (MCP) servers in real time.

## Folder Structure

```
hs-mcp-manager-core-backend/
├── Dockerfile          # Minikube container build file
├── go.mod              # Go module dependencies
├── k8s-deployment.yaml # Kubernetes manifests (PVC, Deployment, Service)
├── main.go             # The core Go orchestration backend & WebSocket engine
├── models.go           # Database structural mappings
├── README.md           # Project instructions
└── schema.sql          # Database initialization script
```

## Minikube Deployment

1. Point your terminal's Docker CLI to Minikube's built-in container engine
```bash
$ eval $(minikube docker-env)
```

2. Build the development-ready image directly inside Minikube
```bash
docker build -t hs-mcp-manager-core-backend:local .
```

3. Apply the Kubernetes manifests
```bash
kubectl apply -f k8s-deployment.yaml
```
