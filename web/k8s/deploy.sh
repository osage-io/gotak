#!/bin/bash

# GoTAK Tactical Web UI - Kubernetes Deployment Script
# Deploys the tactical interface to Kubernetes cluster

set -e

# Configuration
NAMESPACE="gotak"
IMAGE_NAME="gotak-web"
IMAGE_TAG=${1:-"latest"}
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

echo "🚀 Deploying GoTAK Tactical Web UI to Kubernetes"
echo "Namespace: ${NAMESPACE}"
echo "Image: ${FULL_IMAGE_NAME}"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster"
    exit 1
fi

echo "✅ Connected to cluster: $(kubectl config current-context)"

# Create namespace if it doesn't exist
echo "📁 Creating namespace if it doesn't exist..."
kubectl apply -f namespace.yaml

# Apply ConfigMap first
echo "⚙️  Applying ConfigMap..."
kubectl apply -f configmap.yaml

# Build and load the Docker image (for local development)
if [[ "${IMAGE_TAG}" == "latest" || "${IMAGE_TAG}" == "dev" ]]; then
    echo "🔨 Building Docker image..."
    cd ..
    docker build -t ${FULL_IMAGE_NAME} .
    
    # For local k8s (minikube, kind, etc.), load the image
    if command -v minikube &> /dev/null && minikube status &> /dev/null; then
        echo "📦 Loading image to minikube..."
        minikube image load ${FULL_IMAGE_NAME}
    elif command -v kind &> /dev/null; then
        echo "📦 Loading image to kind..."
        kind load docker-image ${FULL_IMAGE_NAME}
    fi
    cd k8s
fi

# Apply the deployment
echo "🚢 Applying Deployment..."
# Update the image in deployment.yaml
sed -i.bak "s|image: gotak-web:latest|image: ${FULL_IMAGE_NAME}|g" deployment.yaml
kubectl apply -f deployment.yaml
# Restore original file
mv deployment.yaml.bak deployment.yaml

# Apply Service
echo "🌐 Applying Service..."
kubectl apply -f service.yaml

# Apply Ingress (optional)
if [[ "${2}" == "--with-ingress" ]]; then
    echo "🌍 Applying Ingress..."
    kubectl apply -f ingress.yaml
fi

echo "⏳ Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/gotak-web -n ${NAMESPACE}

# Show deployment status
echo "📊 Deployment Status:"
kubectl get pods -n ${NAMESPACE} -l app.kubernetes.io/name=gotak-web
kubectl get services -n ${NAMESPACE} -l app.kubernetes.io/name=gotak-web

# Get access information
echo ""
echo "🎯 Access Information:"
NODE_PORT=$(kubectl get service gotak-web-nodeport -n ${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}')
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    MINIKUBE_IP=$(minikube ip)
    echo "📱 Minikube Access: http://${MINIKUBE_IP}:${NODE_PORT}"
elif kubectl get nodes -o wide &> /dev/null; then
    NODE_IP=$(kubectl get nodes -o wide | awk 'NR==2{print $6}')
    echo "📱 NodePort Access: http://${NODE_IP}:${NODE_PORT}"
fi

if [[ "${2}" == "--with-ingress" ]]; then
    echo "🌐 Ingress Access: https://gotak.local (add to /etc/hosts if needed)"
fi

echo "✅ GoTAK Tactical Web UI deployed successfully!"

# Show logs
echo ""
echo "📋 Recent logs:"
kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=gotak-web --tail=5
