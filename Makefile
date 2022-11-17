.PHONY: all

all:
	docker-compose up -d --build

clean:
	docker-compose down --volumes

restart:
	docker-compose down --volumes
	docker-compose up -d --build