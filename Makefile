PKG := gitlab.com/blender/flamenco-ng-poc
VERSION := $(shell git describe --tags --dirty --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

LDFLAGS := -X ${PKG}/internal/appinfo.ApplicationVersion=${VERSION}
BUILD_FLAGS = -ldflags="${LDFLAGS}"

export CGO_ENABLED=0

all: application

# Install generators and build the software.
with-deps:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
	go get github.com/golang/mock/mockgen@v1.6.0
	$(MAKE) application

application: ${RESOURCES} generate flamenco-manager-poc flamenco-worker-poc socketio-poc

flamenco-manager-poc:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager-poc

flamenco-worker-poc:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-worker-poc

socketio-poc:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/socketio-poc

generate:
	go generate ./pkg/api/...
	go generate ./internal/...

# resource.syso: resource/thermogui.ico resource/versioninfo.json
# 	goversioninfo -icon=resource/thermogui.ico -64 resource/versioninfo.json

version:
	@echo "OS     : ${OS}"
	@echo "Package: ${PKG}"
	@echo "Version: ${VERSION}"
	@echo
	@env | grep GO

list-embedded:
	@go list -f '{{printf "%10s" .Name}}: {{.EmbedFiles}}' ${PKG}/...

swagger-ui:
	git clone --depth 1 https://github.com/swagger-api/swagger-ui.git tmp-swagger-ui
	rm -rf pkg/api/static/swagger-ui
	mv tmp-swagger-ui/dist pkg/api/static/swagger-ui
	rm -rf tmp-swagger-ui
	@echo
	@echo 'Now update pkg/api/static/swagger-ui/index.html to have url: "/api/openapi3.json",'

test:
	go test -p 1 -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

clean:
	@go clean -i -x
	rm -f flamenco*-poc-v* flamenco*-poc *.exe resource.syso
	rm -f pkg/api/*.gen.go internal/*/mocks/*.gen.go internal/*/*/mocks/*.gen.go
	@$(MAKE) generate

static: vet lint generate
	go build -v -o flamenco-manager-poc-static -tags netgo -ldflags="-extldflags \"-static\" -w -s ${LDFLAGS}" ${PKG}/cmd/flamenco-manager-poc
	go build -v -o flamenco-worker-poc-static -tags netgo -ldflags="-extldflags \"-static\" -w -s ${LDFLAGS}" ${PKG}/cmd/flamenco-worker-poc

.PHONY: run application version static vet lint deploy  flamenco-manager flamenco-worker socketio-poc generate with-deps
