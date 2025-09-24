# Keycloak User Management Script

## Pre-requisites

1. Ensure the following environment variables are set:
   - `BASE_URL`: The base URL of the Keycloak server (e.g., `https://auth.example.com`).
   - `CLIENT_ID`: The client ID for authentication. Defaults to `user-manage-client` if not provided.
   - `CLIENT_SECRET`: The client secret for authentication.

2. Install `curl` on your system if not already available.

## Usage

This script allows enabling or disabling a Keycloak user.

### Enable a User
```bash
./kc_user.sh --enable <username>
```

### Disable a User
```bash
./kc_user.sh --disable <username>
```

Replace `<username>` with the username of the Keycloak user you want to manage.
