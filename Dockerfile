# Multi-stage build: Frontend + Backend in one image

# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend (creates dist/ folder with static files)
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

# Copy backend go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go

# Stage 3: Runtime - Single image with both
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/server ./server

# Copy frontend static files
COPY --from=frontend-builder /app/frontend/dist ./public

# Expose single port
EXPOSE 8080

# Run the backend server (which will serve frontend static files)
CMD ["./server"]
