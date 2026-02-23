#!/bin/bash

set -e

echo "================================"
echo "Installing Chaos Mesh"
echo "================================"
echo ""

# Check prerequisites
if ! command -v helm &> /dev/null; then
    echo "❌ helm not found. Install from: https://helm.sh/docs/intro/install/"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Install from: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

# Add Chaos Mesh Helm repo
echo "[1/3] Adding Chaos Mesh Helm repository..."
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update
echo "✅ Repository added"
echo ""

# Create namespace
echo "[2/3] Creating chaos-mesh namespace..."
kubectl create namespace chaos-mesh --dry-run=client -o yaml | kubectl apply -f -
echo "✅ Namespace created"
echo ""

# Install Chaos Mesh
echo "[3/3] Installing Chaos Mesh..."
helm install chaos-mesh chaos-mesh/chaos-mesh \
  -n chaos-mesh \
  --set chaosDaemon.runtime=docker \
  --set chaosDaemon.socketPath=/var/run/docker.sock \
  --set dashboard.enabled=true \
  --set dashboard.service.type=NodePort \
  --wait

echo "✅ Chaos Mesh installed"
echo ""

# Wait for pods
echo "Waiting for Chaos Mesh pods to be ready..."
kubectl rollout status deployment/chaos-controller-manager -n chaos-mesh --timeout=5m || true
kubectl rollout status deployment/chaos-dashboard -n chaos-mesh --timeout=5m || true
kubectl rollout status daemonset/chaos-daemon -n chaos-mesh --timeout=5m || true

echo ""
echo "================================"
echo "✅ Chaos Mesh Installation Complete!"
echo "================================"
echo ""
echo "Verify installation:"
echo "  kubectl get pods -n chaos-mesh"
echo ""
echo "Access Dashboard:"
echo "  kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333"
echo "  Open: http://localhost:2333"
echo ""
echo "Next: bash ./run-experiments.sh"
echo ""
