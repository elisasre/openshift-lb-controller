CONTROLLER_NAME  := openshift-lb-controller
IMAGE := yourregistry/$(CONTROLLER_NAME)
.PHONY: install_deps build build-image test

test:
	go test github.com/ElisaOyj/openshift-lb-controller/pkg/controller
	golint -set_exit_status cmd/... && golint -set_exit_status pkg/...
	./.travis.gofmt.sh

install_deps:
	dep ensure -v

build:
	rm -rf bin/$(CONTROLLER_NAME)
	CGO_ENABLED=0 go build -v -o bin/$(CONTROLLER_NAME) ./cmd

build-image:
	rm -rf $(CONTROLLER_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(CONTROLLER_NAME) ./cmd
	docker build -t $(IMAGE):latest .
	docker push $(IMAGE):latest
