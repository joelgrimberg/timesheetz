#!/bin/bash
#
# Timesheetz PostgreSQL Kubernetes Setup Script
#
# This script creates Kubernetes manifests for deploying PostgreSQL
# for use with Timesheetz sync functionality.
#
# Usage:
#   ./setup-postgres-k8s.sh [OPTIONS]
#
# Options:
#   -d, --dir DIR           Directory to create manifest files (default: ./timesheetz-k8s)
#   -n, --namespace NS      Kubernetes namespace (default: timesheetz)
#   --password PASS         Use specific password (default: auto-generated)
#   --storage-class CLASS   Storage class for PVC (default: standard)
#   --storage-size SIZE     Storage size (default: 5Gi)
#   --help                  Show this help message

set -e

# Default values
INSTALL_DIR="./timesheetz-k8s"
NAMESPACE="timesheetz"
PG_PASSWORD=""
STORAGE_CLASS="standard"
STORAGE_SIZE="5Gi"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[OK]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
    sed -n '3,17p' "$0" | sed 's/^# //' | sed 's/^#//'
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --password)
            PG_PASSWORD="$2"
            shift 2
            ;;
        --storage-class)
            STORAGE_CLASS="$2"
            shift 2
            ;;
        --storage-size)
            STORAGE_SIZE="$2"
            shift 2
            ;;
        --help)
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Generate password if not provided
if [ -z "$PG_PASSWORD" ]; then
    if command -v openssl &> /dev/null; then
        PG_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
    else
        PG_PASSWORD=$(head -c 32 /dev/urandom | base64 | tr -d '/+=' | head -c 32)
    fi
fi

# Base64 encode for Kubernetes secret
PG_PASSWORD_B64=$(echo -n "$PG_PASSWORD" | base64)
PG_USER_B64=$(echo -n "timesheetz" | base64)
PG_DB_B64=$(echo -n "timesheetz" | base64)

print_info "Timesheetz PostgreSQL Kubernetes Setup"
echo ""

# Create directory
print_info "Creating directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Create namespace manifest
cat > "$INSTALL_DIR/00-namespace.yaml" << EOF
apiVersion: v1
kind: Namespace
metadata:
  name: ${NAMESPACE}
  labels:
    app: timesheetz
EOF
print_success "Created 00-namespace.yaml"

# Create secret manifest
cat > "$INSTALL_DIR/01-secret.yaml" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: timesheetz-postgres-secret
  namespace: ${NAMESPACE}
  labels:
    app: timesheetz
    component: postgres
type: Opaque
data:
  POSTGRES_USER: ${PG_USER_B64}
  POSTGRES_PASSWORD: ${PG_PASSWORD_B64}
  POSTGRES_DB: ${PG_DB_B64}
EOF
print_success "Created 01-secret.yaml"

# Create PVC manifest
cat > "$INSTALL_DIR/02-pvc.yaml" << EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: timesheetz-postgres-pvc
  namespace: ${NAMESPACE}
  labels:
    app: timesheetz
    component: postgres
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ${STORAGE_CLASS}
  resources:
    requests:
      storage: ${STORAGE_SIZE}
EOF
print_success "Created 02-pvc.yaml"

# Create deployment manifest
cat > "$INSTALL_DIR/03-deployment.yaml" << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: timesheetz-postgres
  namespace: ${NAMESPACE}
  labels:
    app: timesheetz
    component: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: timesheetz
      component: postgres
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: timesheetz
        component: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
              name: postgres
          envFrom:
            - secretRef:
                name: timesheetz-postgres-secret
          volumeMounts:
            - name: postgres-data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              memory: "256Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - timesheetz
                - -d
                - timesheetz
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - timesheetz
                - -d
                - timesheetz
            initialDelaySeconds: 5
            periodSeconds: 5
      volumes:
        - name: postgres-data
          persistentVolumeClaim:
            claimName: timesheetz-postgres-pvc
EOF
print_success "Created 03-deployment.yaml"

# Create service manifest
cat > "$INSTALL_DIR/04-service.yaml" << EOF
apiVersion: v1
kind: Service
metadata:
  name: timesheetz-postgres
  namespace: ${NAMESPACE}
  labels:
    app: timesheetz
    component: postgres
spec:
  type: ClusterIP
  ports:
    - port: 5432
      targetPort: 5432
      protocol: TCP
      name: postgres
  selector:
    app: timesheetz
    component: postgres
---
# NodePort service for external access (optional)
# Uncomment if you need to access from outside the cluster
# apiVersion: v1
# kind: Service
# metadata:
#   name: timesheetz-postgres-external
#   namespace: ${NAMESPACE}
#   labels:
#     app: timesheetz
#     component: postgres
# spec:
#   type: NodePort
#   ports:
#     - port: 5432
#       targetPort: 5432
#       nodePort: 30432
#       protocol: TCP
#       name: postgres
#   selector:
#     app: timesheetz
#     component: postgres
EOF
print_success "Created 04-service.yaml"

# Create kustomization file
cat > "$INSTALL_DIR/kustomization.yaml" << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: ${NAMESPACE}

resources:
  - 00-namespace.yaml
  - 01-secret.yaml
  - 02-pvc.yaml
  - 03-deployment.yaml
  - 04-service.yaml

commonLabels:
  app.kubernetes.io/name: timesheetz
  app.kubernetes.io/component: postgres
EOF
print_success "Created kustomization.yaml"

# Create credentials file (not part of manifests, just for reference)
cat > "$INSTALL_DIR/credentials.txt" << EOF
# Timesheetz PostgreSQL Credentials
# Generated: $(date)
# KEEP THIS FILE SECURE!

Namespace: ${NAMESPACE}
Database: timesheetz
Username: timesheetz
Password: ${PG_PASSWORD}

# Internal cluster URL (from within the cluster):
postgres://timesheetz:${PG_PASSWORD}@timesheetz-postgres.${NAMESPACE}.svc.cluster.local:5432/timesheetz?sslmode=disable

# If using NodePort (uncomment in 04-service.yaml):
# postgres://timesheetz:${PG_PASSWORD}@<NODE_IP>:30432/timesheetz?sslmode=disable
EOF
print_success "Created credentials.txt"

# Create README
cat > "$INSTALL_DIR/README.md" << 'READMEEOF'
# Timesheetz PostgreSQL on Kubernetes

## Quick Start

```bash
# Apply all manifests
kubectl apply -k .

# Or apply individually
kubectl apply -f 00-namespace.yaml
kubectl apply -f 01-secret.yaml
kubectl apply -f 02-pvc.yaml
kubectl apply -f 03-deployment.yaml
kubectl apply -f 04-service.yaml

# Check status
kubectl get pods -n timesheetz
kubectl get svc -n timesheetz
```

## Access from Outside Cluster

### Option 1: Port Forward (for testing)
```bash
kubectl port-forward -n timesheetz svc/timesheetz-postgres 5432:5432
```
Then connect to `localhost:5432`

### Option 2: NodePort
Uncomment the NodePort service in `04-service.yaml` and reapply.
Then connect to `<any-node-ip>:30432`

### Option 3: LoadBalancer (cloud providers)
Change service type to `LoadBalancer` in `04-service.yaml`

## Configuration for Timesheetz

Add to `~/.config/timesheetz/config.yaml`:

```yaml
postgresURL: "<see credentials.txt for URL>"
```

## Backup

```bash
kubectl exec -n timesheetz deploy/timesheetz-postgres -- \
  pg_dump -U timesheetz timesheetz > backup.sql
```

## Restore

```bash
cat backup.sql | kubectl exec -i -n timesheetz deploy/timesheetz-postgres -- \
  psql -U timesheetz timesheetz
```

## Delete Everything

```bash
kubectl delete -k .
```
READMEEOF
print_success "Created README.md"

echo ""
echo "=============================================="
echo -e "${GREEN}Kubernetes Manifests Created!${NC}"
echo "=============================================="
echo ""
echo "Files created in: $INSTALL_DIR"
echo ""
echo "  00-namespace.yaml    - Namespace"
echo "  01-secret.yaml       - Database credentials"
echo "  02-pvc.yaml          - Persistent storage"
echo "  03-deployment.yaml   - PostgreSQL deployment"
echo "  04-service.yaml      - Service (ClusterIP)"
echo "  kustomization.yaml   - Kustomize config"
echo "  credentials.txt      - Connection details (KEEP SECURE!)"
echo "  README.md            - Usage instructions"
echo ""
echo "To deploy:"
echo -e "  ${BLUE}kubectl apply -k $INSTALL_DIR${NC}"
echo ""
echo "To access from your machine (port-forward):"
echo -e "  ${BLUE}kubectl port-forward -n $NAMESPACE svc/timesheetz-postgres 5432:5432${NC}"
echo ""
echo "Connection URL (see credentials.txt for password):"
echo -e "  ${YELLOW}postgres://timesheetz:***@localhost:5432/timesheetz?sslmode=disable${NC}"
echo ""
print_warning "Remember to keep credentials.txt secure!"
echo ""
