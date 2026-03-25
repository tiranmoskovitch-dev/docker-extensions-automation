# Frontend-only extension (no backend needed)
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY ui/ .
RUN npm ci && npm run build

FROM alpine:3.19
LABEL org.opencontainers.image.title="Port Dashboard"
LABEL org.opencontainers.image.description="Real-time port and service mapper with health checks"
LABEL org.opencontainers.image.vendor="tiranmoskovitch"
LABEL com.docker.desktop.extension.api.version="0.3.4"
LABEL com.docker.desktop.extension.icon="docker.svg"

COPY metadata.json .
COPY docker.svg .
COPY --from=frontend-builder /app/dist /ui
