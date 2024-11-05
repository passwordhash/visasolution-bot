EMAIL ?= iam@it-yaroslav.ru
INTERVAL ?= 5

# On server
run:
	EMAIL=$(EMAIL) INTERVAL=$(INTERVAL) docker compose up -d --build

# ========================================

docker-compose-build:
	docker-compose build

docker-compose-up:
	docker-compose up -d --build

docker-compose-down:
	docker compose down

#run-app:
#	-docker run --rm -d -p=4444:4444 --shm-size=2g selenium/standalone-chrome
#	- docker run -d -p 2020:2020 --name visasolution-bot docker/visasolution-bot:0.0.1 /app/main -email yaroslav215@icloud.com

