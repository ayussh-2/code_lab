dev-server:
	clear && cd backend && air
dev-frontend:
	clear && cd frontend && bun dev
dev-judge:
	clear && cd backend && go run ./cmd/judge

nats-image:
	docker build -t codelab/nats:latest -f Dockerfile.nats .

nats-up:
	docker run --name codelab-nats -d -p 4222:4222 -p 8222:8222 codelab/nats:latest

nats-down:
	docker rm -f codelab-nats

sandbox-images:
	docker build -t codelab/sandbox-python:latest -f backend/internal/sandbox/images/Dockerfile.python backend/internal/sandbox/images
	docker build -t codelab/sandbox-node:latest   -f backend/internal/sandbox/images/Dockerfile.node   backend/internal/sandbox/images
	docker build -t codelab/sandbox-cpp:latest    -f backend/internal/sandbox/images/Dockerfile.cplus  backend/internal/sandbox/images

sandbox-smoke:
	cd backend && go run ./cmd/sandbox-smoke

seed:
	cd backend && go run ./cmd/seed

seed-reset:
	cd backend && go run ./cmd/seed -reset
