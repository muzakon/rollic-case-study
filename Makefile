build:
	docker compose -f docker-compose.yml up -d --build

up:
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down -v
	docker system prune -f

test-e2e:
	@echo "Running E2E tests against PostgreSQL..."
	@cd server && go test ./tests/e2e/... -v -count=1
