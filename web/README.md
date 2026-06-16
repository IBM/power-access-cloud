# PAC Web UI

Power Access Cloud (PAC) web frontend built with React and Vite.

## Prerequisites

- Node.js (v18 or higher recommended)
- Yarn package manager

## Getting Started

### Install Dependencies

```bash
yarn install
```

## Available Scripts

In the project directory, you can run:

### `yarn dev`

Runs the app in development mode with environment configuration.\
This script sets up environment variables and starts the Vite dev server.

```bash
yarn dev
```

Open [http://localhost:3000](http://localhost:3000) to view it in your browser.

The page will reload when you make changes. You may also see any lint errors in the console.

### `yarn start`

Alternative command to start the development server using env-cmd:

```bash
yarn start
```

### `yarn build`

Builds the app for production to the `dist` folder:

```bash
yarn build
```

It correctly bundles React in production mode and optimizes the build for the best performance.
The build is minified and the filenames include the hashes. Your app is ready to be deployed!

## Technology Stack

- **React 18** - UI library
- **Vite** - Build tool and dev server (fast HMR)
- **Redux Toolkit** - State management
- **Redux Persist** - Persist state to localStorage
- **Axios** - HTTP client for API calls
- **Keycloak** - Authentication and authorization
- **Carbon Design System** - IBM's design system for UI components
- **React Router** - Client-side routing
- **Sass** - CSS preprocessor
- **Yarn** - Package manager

## Package Manager

This project uses **Yarn** as its package manager. All commands should use `yarn` instead of `npm`:

- Install dependencies: `yarn install` or just `yarn`
- Add a package: `yarn add <package-name>`
- Remove a package: `yarn remove <package-name>`
- Run scripts: `yarn <script-name>`

The project includes a `yarn.lock` file which ensures consistent dependency versions across all environments.

## How to use docker compose environment for local development

- Add a host entry for the keycloak container

```bash
$ cat /etc/hosts
127.0.0.1	localhost keycloak
```

- Run the following command to start all containers

```bash
$ docker compose up --build --pull=always
```
