PROJECT_NAME ?= my-app
PKG_MANAGER  ?= pnpm
VERSION      ?= 1.0.0

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  back-run           Run the Go backend (backend/)"
	@echo "  back-test          Run backend tests"
	@echo "  back-lint          Run golangci-lint on the backend"
	@echo "  back-build         Build production backend image (backend/Dockerfile)"
	@echo "  back-deploy        Push the backend image (scripts/sh/deploy.sh)"
	@echo "  back-mocks         Generate backend mocks (pass args as MOCK_ARGS)"
	@echo "  back-swagger       Generate backend Swagger docs (scripts/sh/swagger.sh)"
	@echo "  front-dev          Start the Vite dev server (frontend/)"
	@echo "  front-test         Run frontend tests"
	@echo "  front-lint         Run frontend lint"
	@echo "  front-build        Build the frontend for production (dist/)"
	@echo "  front-build-image  Build production frontend image (frontend/Dockerfile)"
	@echo "  dev                Run backend and frontend together"
	@echo "  test               Run all tests (backend + frontend)"
	@echo "  lint               Run all linters (backend + frontend)"

.PHONY: back-run
back-run:
	@cd backend && go run .

.PHONY: back-test
back-test:
	@cd backend && go test -v -race ./...

.PHONY: back-lint
back-lint:
	@cd backend && golangci-lint run

.PHONY: back-build
back-build:
	@docker build -f backend/Dockerfile \
		--build-arg APP_NAME=$(PROJECT_NAME) \
		-t $(PROJECT_NAME)-api:$(VERSION) backend/

.PHONY: back-deploy
back-deploy:
	@cd backend && scripts/sh/deploy.sh $(PROJECT_NAME)-api $(VERSION)

.PHONY: back-mocks
back-mocks:
	@cd backend && scripts/sh/generate-mocks.sh $(MOCK_ARGS)

.PHONY: back-swagger
back-swagger:
	@cd backend && scripts/sh/swagger.sh

.PHONY: front-dev
front-dev:
	@cd frontend && $(PKG_MANAGER) run dev -- --host

.PHONY: front-test
front-test:
	@cd frontend && $(PKG_MANAGER) run test || echo "No frontend test script defined"

.PHONY: front-lint
front-lint:
	@cd frontend && $(PKG_MANAGER) run lint || echo "No frontend lint script defined"

.PHONY: front-build
front-build:
	@cd frontend && $(PKG_MANAGER) run build

.PHONY: front-build-image
front-build-image:
	@docker build -f frontend/Dockerfile \
		--build-arg PROJECT_NAME=$(PROJECT_NAME) \
		-t $(PROJECT_NAME)-web:$(VERSION) frontend/

.PHONY: dev
dev:
	@$(MAKE) -j2 back-run front-dev

.PHONY: test
test: back-test front-test

.PHONY: lint
lint: back-lint front-lint
