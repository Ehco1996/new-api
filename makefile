FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all build-frontend start-backend

all: start-frontend start-backend

build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm install && DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat VERSION) npm run build npm run build

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

install-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/BurntSushi/toml/cmd/tomlv@master
	go install github.com/go-critic/go-critic/cmd/gocritic@latest

fmt:
	@echo "Running gofmt..."
	pre-commit run --all-files
