FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json ./
COPY yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build

FROM nginxinc/nginx-unprivileged:1.19-alpine AS server

# Nginx config
RUN rm -rf /etc/nginx/conf.d
COPY --chown=nginx:nginx conf /etc/nginx


# Static build
COPY --from=builder --chown=nginx:nginx /app/build /usr/share/nginx/html/

# Default port exposure
EXPOSE 80

WORKDIR /usr/share/nginx/html
COPY --chown=nginx:nginx ./env.sh .
COPY --chown=nginx:nginx .env .
COPY --chown=nginx:nginx ./entrypoint.sh .
USER root
# Add bash
RUN apk add --no-cache bash

RUN chown -R nginx:nginx /usr/share/nginx/html
USER nginx

# Start Nginx server
CMD ["/bin/bash", "/usr/share/nginx/html/entrypoint.sh"]
