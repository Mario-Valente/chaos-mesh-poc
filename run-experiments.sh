#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="chaos-poc"

echo "================================"
echo "Chaos Mesh - Run Experiments"
echo "================================"
echo ""

# Function to run experiment
run_experiment() {
    local name=$1
    local file=$2
    local duration=$3

    echo "Starting experiment: $name"
    echo "Duration: $duration"
    echo "File: $file"
    echo ""

    kubectl apply -f "$file"

    echo "✅ Experiment applied"
    echo "Monitor with:"
    echo "  kubectl get chaosexperiment -n ${NAMESPACE}"
    echo "  kubectl describe chaosexperiment -n ${NAMESPACE}"
    echo ""

    # Wait for specified duration
    echo "Running for ${duration}... (Press Ctrl+C to stop early)"
    sleep "$duration"

    echo ""
    echo "Stopping experiment..."
    kubectl delete -f "$file" 2>/dev/null || true
    echo "✅ Experiment stopped"
    echo ""
}

# Check if chaos mesh is installed
echo "[Check] Verifying Chaos Mesh installation..."
if ! kubectl get namespace chaos-mesh &> /dev/null; then
    echo "❌ Chaos Mesh not found. Run: bash ./install-chaos-mesh.sh"
    exit 1
fi
echo "✅ Chaos Mesh found"
echo ""

# List available experiments
echo "Available experiments:"
echo "  1. Network Chaos (Latency, Packet Loss, Bandwidth, Partition)"
echo "  2. Resource Chaos (CPU, Memory, Disk I/O Stress)"
echo "  3. Pod Chaos (Pod Kill, Pod Failure, Container Kill)"
echo "  4. Run All Experiments (Sequential)"
echo ""

read -p "Select experiment (1-4): " choice

case $choice in
    1)
        echo "Selected: Network Chaos"
        echo ""
        run_experiment "Backend Latency" "chaos-experiments/network-chaos.yaml" "5m"
        ;;
    2)
        echo "Selected: Resource Chaos"
        echo ""
        run_experiment "CPU and Memory Stress" "chaos-experiments/resource-chaos.yaml" "5m"
        ;;
    3)
        echo "Selected: Pod Chaos"
        echo ""
        run_experiment "Pod Failures" "chaos-experiments/pod-chaos.yaml" "3m"
        ;;
    4)
        echo "Selected: Run All Experiments (Sequential)"
        echo ""

        echo "[1/3] Network Chaos"
        run_experiment "Network Chaos" "chaos-experiments/network-chaos.yaml" "5m"

        echo "[2/3] Resource Chaos"
        run_experiment "Resource Chaos" "chaos-experiments/resource-chaos.yaml" "5m"

        echo "[3/3] Pod Chaos"
        run_experiment "Pod Chaos" "chaos-experiments/pod-chaos.yaml" "3m"

        echo "✅ All experiments completed"
        ;;
    *)
        echo "❌ Invalid selection"
        exit 1
        ;;
esac

echo ""
echo "================================"
echo "Experiment Session Complete"
echo "================================"
echo ""
echo "To monitor logs:"
echo "  kubectl logs -n ${NAMESPACE} -f deployment/frontend"
echo "  kubectl logs -n ${NAMESPACE} -f deployment/backend"
echo "  kubectl logs -n ${NAMESPACE} -f deployment/postgres"
echo ""
echo "To check metrics:"
echo "  kubectl get pods -n ${NAMESPACE}"
echo "  kubectl top pods -n ${NAMESPACE}"
echo ""
