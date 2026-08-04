kind create cluster --name pac-test
make -C ../api generate
make -C ../api install
kind get kubeconfig -n pac-test > /tmp/kubeconfig
podman run -p 27017:27017 -d --name mongo -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=dummypasswd mongo:latest
podman build -f ../api/Dockerfile -t e2e-api
podman build --platform linux/arm64 -f ../web/Dockerfile.dev -t e2e-web

#podman run -p 8000:8000 -e KEYCLOAK_CLIENT_SECRET="blx8wHOCwnwvwA1sSolpFsO2eM1jPf5X" -e KEYCLOAK_CLIENT_ID="pac-go-server" -e KEYCLOAK_REALM="pac" -e KEYCLOAK_HOSTNAME="http://localhost:8080" -e KEYCLOAK_SERVICE_ACCOUNT="ui1" -e MONGODB_URI="mongodb://root:dummypasswd@localhost:27017"  --name pac-api --network host -d localhost/e2e-api /pac-go-server
podman run -p 8000:8000 -v ./kubeconfig:/root/.kube/config:ro -e KEYCLOAK_CLIENT_SECRET="blx8wHOCwnwvwA1sSolpFsO2eM1jPf5X" -e KEYCLOAK_CLIENT_ID="pac-go-server" -e KEYCLOAK_REALM="pac" -e KEYCLOAK_HOSTNAME="http://host.containers.internal:8080" -e KEYCLOAK_SERVICE_ACCOUNT="ui1" -e MONGODB_URI="mongodb://root:dummypasswd@host.containers.internal:27017"  --name pac-api -d localhost/e2e-api /pac-go-server

cp ./helpers/.env ../web
podman run --name -e REACT_APP_KEYCLOAK_URL="http://host.containers.internal:8080" -e REACT_APP_KEYCLOAK_REALM="pac" -e REACT_APP_KEYCLOAK_CLIENT_ID="pac-ui" -e REACT_APP_PAC_GO_SERVER_TARGET="http://host.containers.internal:8000" --name pac-web pac-web -p 3000:3000 -d localhost/e2e-web
#podman run -e REACT_APP_KEYCLOAK_URL="http://localhost:8080" -e REACT_APP_KEYCLOAK_REALM="pac" -e REACT_APP_KEYCLOAK_CLIENT_ID="pac-ui" -e REACT_APP_PAC_GO_SERVER_TARGET="http://localhost:8000" --name pac-web --network host -d localhost/e2e-web
