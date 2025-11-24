# IBM® Power® Access Cloud Dev Setup Guide
This guide provides step-by-step instructions for setting up and running the IBM® Power® Access Cloud locally.

## Prerequisites

* Git
* Go
* Node.js and npm
* Yarn package manager

### Part 1: Keycloak and MongoDB setup

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

### Part 2: PAC Server Setup

1. Clone Repository

    ```
    git clone https://github.com/IBM/power-access-cloud/api.git
    cd power-access-cloud
    ```

2. Set Up Kubernetes Cluster
    
    Bring up any Kubernetes cluster (minikube, kind, or cloud-based cluster).

3. Generate Manifests and Deploy

    ```
    make generate && make manifests
    make deploy
    ```

4. Configure Environment Variables

    ```
    export KEYCLOAK_CLIENT_ID=your_client_id
    export KEYCLOAK_CLIENT_SECRET=your_client_secret
    export KEYCLOAK_REALM=your_realm
    export KEYCLOAK_HOSTNAME=your_keycloak_hostname
    export KEYCLOAK_SERVICE_ACCOUNT=your_service_account
    export KEYCLOAK_SERVICE_ACCOUNT_PASSWORD=your_service_account_password
    export MONGODB_URI=your_mongodb_connection_string
    ```

5. Start the Backend Server
    
    ```
    go run main.go
    ```

### Part 3: PAC UI Setup

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
    yarn install
    npm start
    ```

    The UI will be available at http://localhost:3000

### Part 4: PAC Controller Setup
1. Create IBM Cloud API Key
2. Configure and Run Controller

    ```
    export IBMCLOUD_APIKEY=your_ibm_cloud_api_key
    cd pac
    make run
    ```    