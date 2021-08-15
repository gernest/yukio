IMG ?= gernest/yukio:dev

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o yukio.linux
	docker build -f Dockerfile.local . -t ${IMG}

up: build
	docker-compose up