dev-server:
	clear && cd backend && air
dev-frontend:
	clear && cd frontend && bun dev

sandbox-images:
	docker build -t codelab/sandbox-python:latest -f backend/internal/sandbox/images/Dockerfile.python backend/internal/sandbox/images
	docker build -t codelab/sandbox-node:latest   -f backend/internal/sandbox/images/Dockerfile.node   backend/internal/sandbox/images
	docker build -t codelab/sandbox-cpp:latest    -f backend/internal/sandbox/images/Dockerfile.cplus  backend/internal/sandbox/images

sandbox-smoke:
	cd backend && go run ./cmd/sandbox-smoke
