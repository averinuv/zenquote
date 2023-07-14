# Building docker images for server, client and redis
docker-build:
	docker-compose -f deployments/docker-compose.yaml build

# Starting docker-compose services
docker-up:
	docker-compose -f deployments/docker-compose.yaml up

# Stopping docker-compose services
docker-down:
	docker-compose -f deployments/docker-compose.yaml down

# docker-compose logs
docker-log:
	docker-compose -f deployments/docker-compose.yaml log

# Testing the code
test:
	go test -v ./...

# Cleaning up the built binaries
clean:
	rm ./bin/server
	rm ./bin/client
