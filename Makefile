.PHONY: build publish login clean compose-up compose-down rebuild reset

DOCKER_COMPOSE = docker compose
IMAGE = stori-challenge
ECR = 280922450508.dkr.ecr.us-east-1.amazonaws.com
PROFILE = personal
REGION = us-east-1

clean:
	go clean
	go mod tidy

build: clean
	docker build -t $(IMAGE) .
	docker tag $(IMAGE):latest $(ECR)/$(IMAGE):latest

publish: build
	docker push $(ECR)/$(IMAGE):latest

login:
	aws ecr get-login-password --region $(REGION) --profile $(PROFILE) | docker login --username AWS --password-stdin $(ECR)

compose-up:
	$(DOCKER_COMPOSE) up -d

compose-down:
	$(DOCKER_COMPOSE) down

rebuild:
	$(DOCKER_COMPOSE) build --no-cache
	$(DOCKER_COMPOSE) up -d --force-recreate

reset:
	$(DOCKER_COMPOSE) down -v
ifeq ($(OS),Windows_NT)
	PowerShell Remove-Item -Recurse -Force C:\docker-data\stori
else
	rm -rf /mnt/c/docker-data/stori
endif
