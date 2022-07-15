PKG := git.blender.org/flamenco
VERSION := $(shell git describe --tags --dirty --always)
# Version used in the OpenAPI-generated code shouldn't contain the '-dirty'
# suffix. In the common development workflow, those files will always be dirty
# (because they're only committed after locally working, which means the
# implementation has already been written).
OAPI_VERSION := $(shell git describe --tags --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

LDFLAGS := -X ${PKG}/internal/appinfo.ApplicationVersion=${VERSION}
BUILD_FLAGS = -ldflags="${LDFLAGS}"

# Package name of the generated Python/JavaScript code for the Flamenco API.
PY_API_PKG_NAME=flamenco.manager
JS_API_PKG_NAME=manager

# The directory that will contain the built webapp files, and some other files
# that will be served as static files by the Flamenco Manager web server.
#
# WARNING: THIS IS USED IN `rm -rf ${WEB_STATIC}`, DO NOT MAKE EMPTY OR SET TO
# ANY ABSOLUTE PATH.
WEB_STATIC=web/static

# Prevent any dependency that requires a C compiler, i.e. only work with pure-Go libraries.
export CGO_ENABLED=0

all: application

# Install generators and build the software.
with-deps:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/gohugoio/hugo@v0.101.0
	$(MAKE) application

application: webapp flamenco-manager flamenco-worker

flamenco-manager:
	$(MAKE) webapp-static
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager

.PHONY: flamenco-manager-without-webapp
flamenco-manager-without-webapp:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager

flamenco-worker:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-worker

.PHONY: stresser
stresser:
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/stresser

addon-packer: cmd/addon-packer/addon-packer.go
	go build -v ${BUILD_FLAGS} ${PKG}/cmd/addon-packer

flamenco-manager_race:
	CGO_ENABLED=1 go build -race -o $@ -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-manager

flamenco-worker_race:
	CGO_ENABLED=1 go build -race -o $@ -v ${BUILD_FLAGS} ${PKG}/cmd/flamenco-worker

webapp:
	yarn --cwd web/app install

webapp-static: addon-packer
	$(MAKE) clean-webapp-static
# When changing the base URL, also update the line
# e.GET("/app/*", echo.WrapHandler(webAppHandler))
# in `cmd/flamenco-manager/main.go`
	yarn --cwd web/app build --outDir ../static --base=/app/
# yarn --cwd web/app build --outDir ../static --base=/app/ --minify false
	./addon-packer -filename ${WEB_STATIC}/flamenco3-addon.zip
	@echo "Web app has been installed into ${WEB_STATIC}"

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
		-i pkg/api/flamenco-openapi.yaml \
		-g python \
		-o addon/ \
		--package-name "${PY_API_PKG_NAME}" \
		--http-user-agent "Flamenco/${OAPI_VERSION} (Blender add-on)" \
		-p generateSourceCodeOnly=true \
		-p projectName=Flamenco \
		-p packageVersion="${OAPI_VERSION}"

# The generator outputs files so that we can write our own tests. We don't,
# though, so it's better to just remove those placeholders.
	rm -rf addon/flamenco/manager/test

generate-js:
# The generator doesn't consistently overwrite existing files, nor does it
# remove no-longer-generated files.
	rm -rf web/app/src/manager-api
	rm -rf web/_tmp-manager-api-javascript

# See https://openapi-generator.tech/docs/generators/javascript for the options.
# Version '0.0.0' is used as NPM doesn't like Git hashes as versions.
#
# -p modelPropertyNaming=original is needed because otherwise the generator will
# use original naming internally, but generate docs with camelCase, and then
# things don't work properly.
	java -jar addon/openapi-generator-cli.jar \
		generate \
		-i pkg/api/flamenco-openapi.yaml \
		-g javascript \
		-o web/_tmp-manager-api-javascript \
		--http-user-agent "Flamenco/${OAPI_VERSION} / webbrowser" \
		-p projectName=flamenco-manager \
		-p projectVersion="0.0.0" \
		-p apiPackage="${JS_API_PKG_NAME}" \
		-p disallowAdditionalPropertiesIfNotPresent=false \
		-p usePromises=true \
		-p moduleName=flamencoManager

# Cherry-pick the generated sources, and remove everything else.
# The only relevant bit is that the generated code depends on `superagent`,
# which is listed in our `.
	mv web/_tmp-manager-api-javascript/src web/app/src/manager-api
	rm -rf web/_tmp-manager-api-javascript

version:
	@echo "OS          : ${OS}"
	@echo "Package     : ${PKG}"
	@echo "Version     : ${VERSION}"
	@echo "OAPI Version: ${OAPI_VERSION}"
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
# Ensure the web-static directory exists, so that `web/web_app.go` can embed something.
	mkdir -p ${WEB_STATIC}
	go test -p 1 -short ${PKG_LIST}

clean:
	@go clean -i -x
	rm -f flamenco*-v* flamenco-manager flamenco-worker *.exe flamenco-*_race addon-packer
	$(MAKE) clean-webapp-static

clean-webapp-static:
# Start with `./` to avoid horrors when WEB_STATIC is absolute (like / or /home/yourname).
	rm -rf ./${WEB_STATIC}
# Make sure there is at least something to embed by Go, or it may cause some errors.
	mkdir -p ./${WEB_STATIC}
	touch ${WEB_STATIC}/emptyfile

site:
	rm -rf web/flamenco-io-site/public/
	cd web/flamenco-io-site; hugo --baseURL https://www.flamenco.io/
	rsync web/flamenco-io-site/public/ flamenco.io:flamenco.io/ \
		-va \
		--exclude v2/ \
		--exclude .well-known/ \
		--exclude .htaccess \
		--delete-after

package: flamenco-manager flamenco-worker addon-packer
	rm -rf dist-build
	mkdir -p dist-build
	cp -a flamenco-manager flamenco-worker dist-build/
	cp -a web/static/flamenco3-addon.zip dist-build/
	cp -a README.md LICENSE dist-build/
	cd dist-build; zip -r -9 flamenco-${VERSION}.zip *
	mkdir -p dist
	mv dist-build/flamenco-${VERSION}.zip dist
	rm -rf dist-build

.PHONY: application version flamenco-manager flamenco-worker flamenco-manager_race flamenco-worker_race webapp webapp-static generate generate-go generate-py with-deps swagger-ui list-embedded test clean clean-webapp-static
