build:
	docker compose -f docker-compose.yml up -d --build

up:
	docker-compose up -d

down:
	docker-compose down

clean:
	docker-compose down -v
	docker system prune -f
