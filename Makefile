
PROTO_FILE = ./api/rpcscout.proto
TAG = v0.0.6
IMAGE = quay.io/lpabon/rpcscout
IMAGETAG = latest

all: proto rpcscout

./bin:
	mkdir ./bin

rpcscout: ./bin
	CGO_ENABLED=0 GOOS=linux \
		go build -ldflags '-extldflags "-static"' \
		-o ./bin/rpcscout cmd/rpcscout.go

clean:
	rm -rf ./bin

container:
	docker build -t $(IMAGE):$(IMAGETAG) .

proto:
	docker run \
		--privileged --rm \
		-v $(shell pwd):/go/src/code \
		-e "GOPATH=/go" \
		-e "DOCKER_PROTO=yes" \
		-e "PROTO_USER=$(shell id -u)" \
		-e "PROTO_GROUP=$(shell id -g)" \
		-e "PATH=/bin:/usr/bin:/usr/local/bin:/go/bin:/usr/local/go/bin" \
		quay.io/openstorage/grpc-framework:$(TAG)\
			make docker-proto

docker-proto:
ifndef DOCKER_PROTO
	$(error Do not run directly. Run 'make proto' instead.)
endif
	grpcfw $(PROTO_FILE)
	grpcfw-rest $(PROTO_FILE)
	grpcfw-doc $(PROTO_FILE)


.PHONY: proto clean docker-proto server client