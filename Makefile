.PHONY: all

all:
	docker-compose up -d --build

clean:
	docker-compose down --volumes

restart:
	docker-compose down --volumes
	docker-compose up -d --build

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

generate_swagger:
	swagger generate spec -o ./api/swagger/swagger.yaml --scan-models --work-dir=./internal/ports/http
	swagger generate spec -o ./internal/ports/http/doc/swagger.json --scan-models --work-dir=./internal/ports/http