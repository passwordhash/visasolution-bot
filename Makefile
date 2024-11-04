run: docker-compose-build docker-compose-up

down: docker-compose-down

dev: run-app

#dev-down:
	#ID=$(docker ps -a | grep "selenium/standalone-chrome" | awk '{print $1}' | head -n 1 ) docker stop $ID

# ========================================

docker-compose-build:
	docker-compose build

docker-compose-up:
	docker-compose up -d --build

docker-compose-down:
	docker compose down

run-app:
	-docker run --rm -d -p=4444:4444 --shm-size=2g selenium/standalone-chrome
	go run ./cmd/bot/main.go

