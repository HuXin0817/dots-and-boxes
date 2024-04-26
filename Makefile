.PHONY: all build serve engine client container tidy pushgit fmt generate

# GO=go
# GO=$$GOROOT/bin/go
GO=/usr/local/go/bin/go

all: build

build:
	cd serve && $(GO) build -o ../bin/serve 
	cd engine && $(GO) build -o ../bin/engine 
	cd client && $(GO) build -o ../bin/client 

tidy:
	go mod tidy
	cd client && go mod tidy
	cd serve && go mod tidy
	cd engine && go mod tidy

serve:
	cd serve && $(GO) run serve -f etc/develop.yaml -h 0.0.0.0:${PORT}

engine:
	cd engine && $(GO) run engine -f develop.yaml

client:
	cd client && $(GO) run client

generate:
	$(GO) generate .

fmt:
	gofmt -d .

container:
	docker-compose up -d --scale engine=5
