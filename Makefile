PKG := git.blender.org/flamenco
VERSION := $(shell git describe --tags --dirty --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

LDFLAGS := -X ${PKG}/internal/appinfo.ApplicationVersion=${VERSION}
BUILD_FLAGS = -ldflags="${LDFLAGS}"

# Package name of the generated Python code for the Flamenco API.
PY_API_PKG_NAME=flamenco.manager

# Prevent any dependency that requires a C compiler, i.e. only work with pure-Go libraries.
export CGO_ENABLED=0

all: application

# Install generators and build the software.
with-deps:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
	go get github.com/golang/mock/mockgen@v1.6.0
	$(MAKE) application

application: flamenco-manager flamenco-worker socketio-poc

flamenco-manager:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager

flamenco-worker:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-worker

socketio-poc:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/socketio-poc

generate: generate-go generate-py

generate-go:
	go generate ./pkg/api/...
	go generate ./internal/...
# The generators always produce UNIX line-ends. This creates false file
# modifications with Git. Convert them to DOS line-ends to avoid this.
ifeq ($(OS),Windows_NT)
	git status --porcelain | grep '^ M .*.gen.go' | cut -d' ' -f3 | xargs unix2dos --keepdate
endif

generate-py:
# The generator doesn't consistently overwrite existing files, nor does it
# remove no-longer-generated files.
	rm -rf addon/flamenco/manager

	java -jar addon/openapi-generator-cli.jar \
		generate \
		-i pkg/api/flamenco-manager.yaml \
		-g python \
		-o addon/ \
		--skip-validate-spec \
		--package-name "${PY_API_PKG_NAME}" \
		--http-user-agent "Flamenco/${VERSION} (Blender add-on)" \
		-p generateSourceCodeOnly=true \
		-p projectName=Flamenco \
		-p packageVersion="${VERSION}"

# The generator outputs files so that we can write our own tests. We don't,
# though, so it's better to just remove those placeholders.
	rm -rf addon/flamenco/manager/test

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

clean:
	@go clean -i -x
	rm -f flamenco*-v* flamenco-manager flamenco-worker socketio-poc *.exe
	rm -f pkg/api/*.gen.go internal/*/mocks/*.gen.go internal/*/*/mocks/*.gen.go
	@$(MAKE) generate

.PHONY: application version flamenco-manager flamenco-worker socketio-poc generate generate-go generate-py with-deps swagger-ui list-embedded test clean
