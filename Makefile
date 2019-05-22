CONTROLLER_NAME  := openshift-lb-controller
IMAGE := elisaoyj/$(CONTROLLER_NAME)
.PHONY: test gofmt check ensure build build-image build-linux-amd64

test:
	go test github.com/ElisaOyj/openshift-lb-controller/pkg/controller
	golint -set_exit_status cmd/... pkg/...
	./hack/gofmt.sh

gofmt:
	./hack/gofmt.sh

check:
	dep check|grep "lock is out of sync"; test $$? -eq 1

ensure:
	dep ensure -v

build:
	rm -rf bin/$(CONTROLLER_NAME)
	CGO_ENABLED=0 go build -v -o bin/$(CONTROLLER_NAME) ./cmd

build-image:
	rm -rf $(CONTROLLER_NAME)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(CONTROLLER_NAME) ./cmd
	docker build -t $(IMAGE):latest .
	docker push $(IMAGE):latest

build-linux-amd64:
	rm -f bin/linux/$(OPERATOR_NAME)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -i -o bin/linux/$(OPERATOR_NAME) ./cmd
