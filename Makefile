build:
	go build -v -o ./bin/app .
.PHONY:build

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.43.0

lint:
	golangci-lint run ./...

docker-build:
	docker build -t io.guthub/zaz600/suomen-botti:latest .

run:
	docker-compose -f docker-compose.yml -p suomen-botti up -d --build

run-log:
	docker-compose -f docker-compose.yml -p suomen-botti up --build

stop:
	docker-compose -f docker-compose.yml -p suomen-botti down

test:
	go fmt ./...
	go vet ./...
	go test -v ./...
	go test -v -race -count 100 ./...
	go test -gcflags=-l -count=1 -timeout=30s -bench=. -run=^$  ./...

	#go test -cover ./... | grep coverage