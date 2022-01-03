OUT := $(notdir $(realpath .))
PKG := stuvel.eu/flamenco-test/goja
VERSION := $(shell git describe --tags --dirty --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
STATIC_OUT := ${OUT}-${VERSION}
PACKAGE_PATH := dist/${OUT}-${VERSION}

LDFLAGS := -X ${PKG}/appinfo.ApplicationVersion=${VERSION}
BUILD_FLAGS = -ldflags="${LDFLAGS}"

ifndef PACKAGE_PATH
# ${PACKAGE_PATH} is used in 'rm' commands, so it's important to check.
$(error PACKAGE_PATH is not set)
endif

RESOURECS :=
ifeq ($(OS),Windows_NT)
	OUT := $(OUT).exe
	STATIC_OUT := $(STATIC_OUT).exe
	LDFLAGS := $(LDFLAGS) -H=windowsgui
	RESOURECS := resource.syso
endif

all: application

application: ${RESOURCES}
	go build -v -o ${OUT} ${BUILD_FLAGS} ${PKG}

install: ${RESOURCES}
	go install -i -v ${BUILD_FLAGS} ${PKG}

resource.syso: resource/thermogui.ico resource/versioninfo.json
	goversioninfo -icon=resource/thermogui.ico -64 resource/versioninfo.json

version:
	@echo "OS     : ${OS}"
	@echo "Package: ${PKG}"
	@echo "Version: ${VERSION}"
	@echo "Target : ${OUT}"

embedded:
	@go list -f "{{.EmbedFiles}}" ${PKG}/job_compilers

test:
	go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

run: application
	./${OUT}

clean:
	@go clean -i -x
	rm -f ${OUT}-v* resource.syso

static: vet lint resource.syso
	go build -v -o ${STATIC_OUT} -tags netgo -ldflags="-extldflags \"-static\" -w -s ${LDFLAGS}" ${PKG}

.gitlabAccessToken:
	$(error gitlabAccessToken does not exist, visit Visit https://gitlab.com/profile/personal_access_tokens, create a Personal Access Token with API access then save it to the file .gitlabAccessToken)


release: .gitlabAccessToken package
	rsync ${PACKAGE_PATH}* stuvel@stuvel.eu:files/beatstripper/ -va
	go run release/release.go -version ${VERSION} -fileglob ${PACKAGE_PATH}\*


package:
	@$(MAKE) _prepare_package
	@$(MAKE) _package_linux
	@$(MAKE) _package_windows
	#@$(MAKE) _package_darwin
	@$(MAKE) _finish_package

package_linux:
	@$(MAKE) _prepare_package
	@$(MAKE) _package_linux
	@$(MAKE) _finish_package

package_windows:
	@$(MAKE) _prepare_package
	@$(MAKE) _package_windows
	@$(MAKE) _finish_package

package_darwin:
	@$(MAKE) _prepare_package
	@$(MAKE) _package_darwin
	@$(MAKE) _finish_package

_package_linux:
	@$(MAKE) --no-print-directory GOOS=linux MONGOOS=linux GOARCH=amd64 STATIC_OUT=${PACKAGE_PATH}/${OUT} _package_tar

_package_windows:
	@$(MAKE) --no-print-directory GOOS=windows MONGOOS=windows GOARCH=amd64 STATIC_OUT=${PACKAGE_PATH}/${OUT}.exe _package_zip

_package_darwin:
	@$(MAKE) --no-print-directory GOOS=darwin MONGOOS=osx GOARCH=amd64 STATIC_OUT=${PACKAGE_PATH}/${OUT} _package_zip

_prepare_package:
	rm -rf ${PACKAGE_PATH}
	mkdir -p ${PACKAGE_PATH}
	cp -ua README.md LICENSE ${PACKAGE_PATH}/

_finish_package:
	rm -r ${PACKAGE_PATH}
	rm -f ${PACKAGE_PATH}.sha256
	sha256sum ${PACKAGE_PATH}* | tee ${PACKAGE_PATH}.sha256

_package_tar: static
	tar -C $(dir ${PACKAGE_PATH}) -zcf $(PWD)/${PACKAGE_PATH}-${GOOS}.tar.gz $(notdir ${PACKAGE_PATH})
	rm ${STATIC_OUT}

_package_zip: static
	cd $(dir ${PACKAGE_PATH}) && zip -9 -r -q $(notdir ${PACKAGE_PATH})-${GOOS}.zip $(notdir ${PACKAGE_PATH})
	rm ${STATIC_OUT}

.PHONY: run application version static vet lint deploy package release
