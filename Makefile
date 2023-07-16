# Building docker images for server, client and redisdb
docker-build:
	docker-compose -f deployments/docker-compose.yaml build

# Starting docker-compose services
docker-up:
	docker-compose -f deployments/docker-compose.yaml up -d

# Stopping docker-compose services
docker-down:
	docker-compose -f deployments/docker-compose.yaml down

# Server logs
docker-server-log:
	docker-compose -f deployments/docker-compose.yaml logs server

# Client logs
docker-client-log:
	docker-compose -f deployments/docker-compose.yaml logs client

docker-log:
	make docker-server-log
	make docker-client-log

# Testing the code
test:
	go test -v ./...

# Cleaning up the built binaries
clean:
	rm ./bin/server
	rm ./bin/client

proto:
	docker run -v $(PWD):/defs namely/protoc-all -f api/api.proto -l go -o . --go-source-relative
