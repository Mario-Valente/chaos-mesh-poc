# Setup Guide - Chaos Mesh POC

Guia passo a passo para colocar a POC em funcionamento.

## Pré-requisitos

### Instalação obrigatória

1. **Docker**
   ```bash
   # macOS com Homebrew
   brew install docker
   # ou download em: https://www.docker.com/products/docker-desktop
   ```

2. **Kind** (Kubernetes local)
   ```bash
   brew install kind
   # ou: https://kind.sigs.k8s.io/docs/user/quick-start/
   ```

3. **kubectl** (CLI Kubernetes)
   ```bash
   brew install kubectl
   # ou: https://kubernetes.io/docs/tasks/tools/
   ```

4. **Helm** (Package manager Kubernetes)
   ```bash
   brew install helm
   # ou: https://helm.sh/docs/intro/install/
   ```

### Verificação

```bash
docker --version
kind --version
kubectl version --client
helm version
```

## Setup Automático (Recomendado)

O script `setup.sh` automatiza todo o processo:

```bash
cd chaos-mesh-poc
bash setup.sh
```

Isso vai:
1. ✅ Verificar pré-requisitos
2. ✅ Criar cluster Kind
3. ✅ Criar namespace `chaos-poc`
4. ✅ Buildar imagens Docker
5. ✅ Fazer deploy dos serviços

**Tempo estimado**: 3-5 minutos (dependendo da conexão)

## Setup Manual (Passo a Passo)

Se preferir fazer manualmente:

### 1. Criar o cluster Kind

```bash
kind create cluster --name chaos-mesh-poc --wait 5m
```

Verificar:
```bash
kubectl cluster-info
kubectl get nodes
```

### 2. Criar namespace

```bash
kubectl create namespace chaos-poc
```

### 3. Buildar imagens Docker

```bash
# Frontend
docker build -f docker/frontend.Dockerfile -t localhost:5000/frontend:latest .

# Backend
docker build -f docker/backend.Dockerfile -t localhost:5000/backend:latest .
```

### 4. Carregar imagens no Kind

```bash
kind load docker-image localhost:5000/frontend:latest --name chaos-mesh-poc
kind load docker-image localhost:5000/backend:latest --name chaos-mesh-poc
```

### 5. Deploy dos serviços

```bash
# Aplicar manifiestos em ordem
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/services.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml
```

### 6. Aguardar deployment

```bash
kubectl rollout status deployment/postgres -n chaos-poc --timeout=5m
kubectl rollout status deployment/backend -n chaos-poc --timeout=5m
kubectl rollout status deployment/frontend -n chaos-poc --timeout=5m
```

## Instalação do Chaos Mesh

Após o setup dos serviços:

```bash
bash install-chaos-mesh.sh
```

Ou manualmente:

```bash
# Adicionar repo
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update

# Criar namespace
kubectl create namespace chaos-mesh

# Instalar
helm install chaos-mesh chaos-mesh/chaos-mesh \
  -n chaos-mesh \
  --set chaosDaemon.runtime=docker \
  --set chaosDaemon.socketPath=/var/run/docker.sock \
  --set dashboard.enabled=true \
  --set dashboard.service.type=NodePort \
  --wait
```

Verificar instalação:
```bash
kubectl get pods -n chaos-mesh
```

## Verificação

Todos os componentes devem estar rodando:

```bash
# Serviços da POC
kubectl get pods -n chaos-poc
# Esperado: frontend (2), backend (2), postgres (1) - todos Running

# Chaos Mesh
kubectl get pods -n chaos-mesh
# Esperado: controller-manager, daemon, dashboard - todos Running
```

## Acessar os Serviços

### Frontend API

```bash
kubectl port-forward -n chaos-poc svc/frontend-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/data
```

### Logs

```bash
# Streaming de logs
kubectl logs -n chaos-poc -f deployment/frontend
kubectl logs -n chaos-poc -f deployment/backend

# Últimas 100 linhas
kubectl logs -n chaos-poc -p deployment/backend --tail=100
```

### Dashboard do Chaos Mesh

```bash
kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333
# Abra: http://localhost:2333
```

## Troubleshooting

### Pods não estão Ready

```bash
# Ver eventos
kubectl describe pod <pod-name> -n chaos-poc

# Ver logs
kubectl logs <pod-name> -n chaos-poc --previous

# Reiniciar pod
kubectl delete pod <pod-name> -n chaos-poc
```

### Problema ao conectar Backend ao PostgreSQL

```bash
# Verificar conexão
kubectl exec -it deployment/backend -n chaos-poc -- \
  psql -h postgres-service -U postgres -d chaos_db -c "SELECT 1"

# Verificar service
kubectl get svc -n chaos-poc
kubectl describe svc postgres-service -n chaos-poc
```

### Imagens Docker não encontradas

```bash
# Verificar imagens no Kind
docker exec -it chaos-mesh-poc-control-plane \
  crictl images

# Recarregar se necessário
kind load docker-image localhost:5000/frontend:latest --name chaos-mesh-poc
```

### Helm não encontra chart

```bash
helm repo update
helm search repo chaos-mesh
```

## Limpeza

### Remover tudo

```bash
# Remover cluster
kind delete cluster --name chaos-mesh-poc

# Remover imagens Docker
docker rmi localhost:5000/frontend:latest
docker rmi localhost:5000/backend:latest
```

### Apenas reset dos serviços

```bash
kubectl delete namespace chaos-poc
kubectl create namespace chaos-poc
# Redeploy dos serviços com kubectl apply
```

## Próximos Passos

1. ✅ Verificar que todos os pods estão Running
2. ✅ Testar a API: `curl http://localhost:8080/api/data`
3. ✅ Acessar o dashboard do Chaos Mesh
4. ✅ Executar experimentos: `bash run-experiments.sh`

Veja [Experiments](experiments.md) para detalhes dos experiments disponíveis.
