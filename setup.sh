#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="chaos-mesh-poc"
NAMESPACE="chaos-poc"
CLUSTER_NAME="chaos-mesh-poc"

echo "================================"
echo "Chaos Mesh POC - Full Setup"
echo "================================"
echo ""

# Check prerequisites
echo "[1/5] Checking prerequisites..."
if ! command -v kind &> /dev/null; then
    echo "❌ kind not found. Install from: https://kind.sigs.k8s.io/docs/user/quick-start/"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Install from: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

if ! command -v helm &> /dev/null; then
    echo "❌ helm not found. Install from: https://helm.sh/docs/intro/install/"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ docker not found. Install from: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "✅ All prerequisites found"
echo ""

# Create Kind cluster
echo "[2/5] Creating Kind cluster..."
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "⚠️  Cluster ${CLUSTER_NAME} already exists, skipping creation"
else
    kind create cluster --name ${CLUSTER_NAME} --wait 5m
    echo "✅ Cluster created"
fi
echo ""

# Create namespace
echo "[3/5] Creating namespace..."
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
echo "✅ Namespace created"
echo ""

# Build and load Docker images
echo "[4/5] Building and loading Docker images..."
cd ${SCRIPT_DIR}

echo "  - Building frontend..."
docker build -f docker/frontend.Dockerfile -t localhost:5000/frontend:latest .

echo "  - Building backend..."
docker build -f docker/backend.Dockerfile -t localhost:5000/backend:latest .

echo "  - Loading images into Kind cluster..."
kind load docker-image localhost:5000/frontend:latest --name ${CLUSTER_NAME}
kind load docker-image localhost:5000/backend:latest --name ${CLUSTER_NAME}

echo "✅ Images built and loaded"
echo ""

# Deploy applications
echo "[5/5] Deploying applications..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/services.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml

echo "✅ Applications deployed"
echo ""

# Wait for deployments
echo "Waiting for deployments to be ready..."
kubectl rollout status deployment/postgres -n ${NAMESPACE} --timeout=5m || true
kubectl rollout status deployment/backend -n ${NAMESPACE} --timeout=5m || true
kubectl rollout status deployment/frontend -n ${NAMESPACE} --timeout=5m || true

echo ""
echo "================================"
echo "✅ Setup Complete!"
echo "================================"
echo ""
echo "Next steps:"
echo "  1. Verify all pods are running:"
echo "     kubectl get pods -n ${NAMESPACE}"
echo ""
echo "  2. Install Chaos Mesh:"
echo "     bash ./install-chaos-mesh.sh"
echo ""
echo "  3. Run chaos experiments:"
echo "     bash ./run-experiments.sh"
echo ""
echo "  4. Monitor the frontend service:"
echo "     kubectl port-forward -n ${NAMESPACE} svc/frontend-service 8080:8080"
echo "     curl http://localhost:8080/api/data"
echo ""
