PKG := git.blender.org/flamenco
VERSION := $(shell git describe --tags --dirty --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

LDFLAGS := -X ${PKG}/internal/appinfo.ApplicationVersion=${VERSION}
BUILD_FLAGS = -ldflags="${LDFLAGS}"

# Package name of the generated Python/JavaScript code for the Flamenco API.
PY_API_PKG_NAME=flamenco.manager
JS_API_PKG_NAME=manager

# Prevent any dependency that requires a C compiler, i.e. only work with pure-Go libraries.
export CGO_ENABLED=0

all: application

# Install generators and build the software.
with-deps:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
	go get github.com/golang/mock/mockgen@v1.6.0
	$(MAKE) application

application: flamenco-manager flamenco-worker

flamenco-manager:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager

flamenco-worker:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-worker

generate: generate-go generate-py generate-js

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

# See https://openapi-generator.tech/docs/generators/python for the options.
	java -jar addon/openapi-generator-cli.jar \
		generate \
		-i pkg/api/flamenco-manager.yaml \
		-g python \
		-o addon/ \
		--package-name "${PY_API_PKG_NAME}" \
		--http-user-agent "Flamenco/${VERSION} (Blender add-on)" \
		-p generateSourceCodeOnly=true \
		-p projectName=Flamenco \
		-p packageVersion="${VERSION}"

# The generator outputs files so that we can write our own tests. We don't,
# though, so it's better to just remove those placeholders.
	rm -rf addon/flamenco/manager/test

generate-js:
# The generator doesn't consistently overwrite existing files, nor does it
# remove no-longer-generated files.
	rm -rf web/manager-api

# See https://openapi-generator.tech/docs/generators/javascript for the options.
# Version '0.0.0' is used as NPM doesn't like Git hashes as versions.
#
# -p modelPropertyNaming=original is needed because otherwise the generator will
# use original naming internally, but generate docs with camelCase, and then
# things don't work properly.
	java -jar addon/openapi-generator-cli.jar \
		generate \
		-i pkg/api/flamenco-manager.yaml \
		-g javascript \
		-o web/manager-api \
		--http-user-agent "Flamenco/${VERSION} / webbrowser" \
		-p projectName=flamenco-manager \
		-p projectVersion="0.0.0" \
		-p apiPackage="${JS_API_PKG_NAME}" \
		-p disallowAdditionalPropertiesIfNotPresent=false \
		-p usePromises=true \
		-p moduleName=flamencoManager \
		-p modelPropertyNaming=original

# The generator outputs files so that we can write our own tests. We don't,
# though, so it's better to just remove those placeholders.
	rm -rf web/manager-api/test

# Ensure that the API package can actually be imported by the webapp (if
# symlinks are in place).
	cd web/manager-api && npm install && npm run build

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
	rm -f flamenco*-v* flamenco-manager flamenco-worker *.exe
	rm -f pkg/api/*.gen.go internal/*/mocks/*.gen.go internal/*/*/mocks/*.gen.go
	@$(MAKE) generate

package: flamenco-manager flamenco-worker
	mkdir -p dist
	rsync -a flamenco-manager flamenco-worker dist/
	rsync -a addon/flamenco dist/ --exclude __pycache__ --exclude '*.pyc' --prune-empty-dirs --exclude .mypy_cache --exclude manager/docs  --delete --delete-excluded
	cd dist; zip -r -9 flamenco-${VERSION}-addon.zip flamenco
	rm -rf dist/flamenco


.PHONY: application version flamenco-manager flamenco-worker generate generate-go generate-py with-deps swagger-ui list-embedded test clean
