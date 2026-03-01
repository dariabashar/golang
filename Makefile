.PHONY: up down logs ps rebuild

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker ps -a

rebuild: down up

