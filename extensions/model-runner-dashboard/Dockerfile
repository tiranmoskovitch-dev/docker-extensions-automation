# Stage 1: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/ .
RUN go mod download && \
    CGO_ENABLED=0 go build -o model-runner-dashboard .

# Stage 2: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY ui/ .
RUN npm ci && npm run build

# Stage 3: Final extension image
FROM alpine:3.19
LABEL org.opencontainers.image.title="Model Runner Dashboard"
LABEL org.opencontainers.image.description="Manage Docker Compose native LLM models and Model Runner"
LABEL org.opencontainers.image.vendor="tiranmoskovitch"
LABEL com.docker.desktop.extension.api.version="0.3.4"
LABEL com.docker.desktop.extension.icon="docker.svg"

COPY metadata.json .
COPY docker.svg .
COPY docker-compose.yaml .
COPY --from=backend-builder /app/model-runner-dashboard /backend/model-runner-dashboard
COPY --from=frontend-builder /app/dist /ui
