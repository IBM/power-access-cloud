# Development

## Theme

### Copy realm-export.json to the relevant folder

```
cp realm-export.json <power-access-cloud-base>/keycloak-infra/keycloak/pac
```

### Start the keycloak via docker compose

```shell
docker compose up
```

### Access the keycloak via

http://localhost:8080

Credentials: `admin/Pa55w0rd`

>Note: any changes done in the theme/pac folder directly visible in the keycloak UI post refresh.

### Cleanup the environment

```shell
docker compose down -v
```
