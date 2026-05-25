# IBM® Power® Access Cloud Dev Setup Guide

This guide provides step-by-step instructions for setting up and running the IBM® Power® Access Cloud locally.

## Prerequisites

* **Git** - Version control
* **Go 1.24**
* **Node.js v18+** - Frontend runtime
* **Yarn** - Package manager (not npm)
* **Kubernetes cluster** - minikube, kind, or cloud-based
* **Podman or Docker** - For running Keycloak and MongoDB

## Architecture Overview

PAC consists of three main components:

1. **PAC Controller** - Kubernetes operator managing Catalog and Service CRDs
2. **PAC Go Server** - REST API backend for UI and integrations
3. **PAC Web UI** - React/Vite frontend with Keycloak authentication

### Part 1: Keycloak and MongoDB Setup

1. From a terminal, enter the following command to start Keycloak:
    ```
    podman run -p 127.0.0.1:8080:8080 -e KC_BOOTSTRAP_ADMIN_USERNAME=admin -e KC_BOOTSTRAP_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:26.3.3 start-dev
    ```
    This command starts Keycloak exposed on the local port 8080 and creates an initial admin user with the username admin and password admin.

2. Enter the following command to start mongodb:
    ```
    podman run -d --name mongo -p 27017:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=root \
    -e MONGO_INITDB_ROOT_PASSWORD=dummypasswd \
    mongo:latest
    ```

### Part 2: PAC Controller Setup (Local Testing)

The PAC Controller is a Kubernetes operator that manages Catalog and Service custom resources.

> **Note:** This setup is for **local development and testing**.
1. **Clone Repository**

    ```bash
    git clone https://github.com/IBM/power-access-cloud.git
    cd power-access-cloud/api
    ```

2. **Set Up Kubernetes Cluster**
    
    Bring up any Kubernetes cluster:
    ```bash
    # Using minikube
    minikube start
    
    # Or using kind
    kind create cluster
    ```

3. **Generate Manifests and Install CRDs**

    ```bash
    cd api
    make generate && make manifests
    make install  # Install CRDs
    ```

### Part 3: PAC Go Server Setup

The PAC Go Server provides REST APIs for the web UI and external integrations.

1. **Navigate to API Directory**

    ```bash
    cd api
    ```

2. **Configure Environment Variables**

    ```bash
    export KEYCLOAK_CLIENT_ID=your_client_id
    export KEYCLOAK_CLIENT_SECRET=your_client_secret
    export KEYCLOAK_REALM=your_realm
    export KEYCLOAK_HOSTNAME=your_keycloak_hostname
    export KEYCLOAK_SERVICE_ACCOUNT=your_service_account
    export KEYCLOAK_SERVICE_ACCOUNT_PASSWORD=your_service_account_password
    export MONGODB_URI=your_mongodb_connection_string
    ```

3. **Build and Run**
    
    ```bash
    # Build all binaries
    make build
    
    # Run pac-go-server
    ./bin/pac-go-server
    
    # Or run directly with go
    go run ./cmd/pac-go-server/main.go
    ```

    The API server will be available at `http://localhost:8080`

4. **View API Documentation**

    Generate Swagger docs:
    ```bash
    make swagger
    ```
    
    Access Swagger UI at: `http://localhost:8080/swagger/index.html`

### Part 4: PAC UI Setup

1. Clone Repository

    ```
    git clone https://github.com/IBM/power-access-cloud.git
    cd power-access-cloud
    ```

2. Configure Environment Variables

    Create a .env file or export the following variables:
    ```
    export REACT_APP_KEYCLOAK_URL=your_keycloak_url
    export REACT_APP_KEYCLOAK_REALM=your_keycloak_realm
    export REACT_APP_KEYCLOAK_CLIENT_ID=your_keycloak_client_id
    export REACT_APP_PAC_GO_SERVER_TARGET=your_backend_server_url
    ```

    and run `./env.sh`

3. Install Dependencies and Run

    ```
    cd web
    yarn install
    yarn dev
    ```

    The UI will be available at http://localhost:3000

## Project Structure

```
api/
├── main.go                    # Controller entry point
├── cmd/
│   ├── pac-go-server/        # REST API server
│   ├── event-notifier/       # Event notification service
│   └── swagger/              # Swagger documentation
├── controllers/              # Kubernetes controllers
│   └── app/                  # App controllers (Catalog, Service)
├── apis/                     # CRD definitions
│   └── app/v1alpha1/        # App CRDs (Catalog, Service, Config)
├── internal/                 # Internal packages
│   └── pkg/pac-go-server/   # API server implementation
└── config/                   # Kubernetes manifests

web/
├── src/                      # React source code
├── public/                   # Static assets
└── package.json             # Dependencies (uses Yarn)
```