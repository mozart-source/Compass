.PHONY: all frontend backend redis stop

SHELL := cmd.exe

# Default target
all: frontend backend redis

# Start Frontend Electron in dev mode
frontend:
	@cmd /c start cmd /k "cd Frontend && npm run electron:dev"

# Start Backend with hot reload using air
backend:
	@cmd /c start cmd /k "cd Backend_go && air"

# Start Redis server in WSL
redis:
	@cmd /c start wsl -e redis-server

# Stop all services (you might need to implement proper process management)
stop:
	@echo "Stopping services..."
	@taskkill /F /IM electron.exe 2>NUL || true
	@taskkill /F /IM main.exe 2>NUL || true
	@wsl -e pkill redis-server 2>NUL || true
	@echo "All services stopped"

# Help target
help:
	@echo "Available targets:"
	@echo "  all      - Start all services (frontend, backend, and redis)"
	@echo "  frontend - Start Frontend Electron in dev mode"
	@echo "  backend  - Start Backend with hot reload"
	@echo "  redis    - Start Redis server in WSL"
	@echo "  stop     - Stop all services" 