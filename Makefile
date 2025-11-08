.PHONY: build publish login clean compose-up compose-down rebuild reset \
        test test-integration test-all \
        tf-init tf-plan tf-apply tf-destroy infra-up infra-down \
        ci

DOCKER_COMPOSE = docker compose
IMAGE = stori-challenge
ECR = 280922450508.dkr.ecr.us-east-1.amazonaws.com
PROFILE = personal
REGION = us-east-1
TF_DIR = deployments/terraform

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

test:
	go test ./internal/... -v -cover

test-integration:
	@echo "Running integration tests..."
	TF_OUTPUT_DB := $(shell terraform -chdir=$(TF_DIR) output -raw db_endpoint 2>/dev/null || echo "localhost")
	DB_HOST=$${TF_OUTPUT_DB} go test ./tests/integration/... -v -tags=integration -count=1

test-all: test test-integration

tf-init:
	cd $(TF_DIR) && terraform init -upgrade

tf-plan:
	cd $(TF_DIR) && terraform plan -var="aws_profile=$(PROFILE)" -var="aws_region=$(REGION)"

tf-apply:
	cd $(TF_DIR) && terraform apply -auto-approve -var="aws_profile=$(PROFILE)" -var="aws_region=$(REGION)"

tf-destroy:
	cd $(TF_DIR) && terraform destroy -auto-approve -var="aws_profile=$(PROFILE)" -var="aws_region=$(REGION)"

infra-up: tf-init tf-apply
infra-down: tf-destroy

ci: login publish test-all infra-up
	@echo "✅ CI/CD pipeline ejecutado correctamente (ECR + Terraform + Tests)"
