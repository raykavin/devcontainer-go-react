#!/bin/sh
# Simple healthcheck used by the production image (backend/Dockerfile).
# Port must match application.health_server_listen_port in config.yml.
wget -q --spider "http://localhost:${HEALTH_SERVER_LISTEN_PORT:-3001}/healthz" || exit 1
