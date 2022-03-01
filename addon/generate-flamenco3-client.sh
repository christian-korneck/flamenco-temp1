#!/bin/bash

# Generator for the Python API client.
#
# See:
# - https://github.com/OpenAPITools/openapi-generator
# - https://openapi-generator.tech/
# - https://openapi-generator.tech/docs/generators/python

PKG_NAME=flamenco.manager
PKG_VERSION=3.0

set -ex

# The generator doesn't consistently overwrite existing files, nor does it
# remove no-longer-generated files.
rm -rf ./flamenco/manager

java -jar openapi-generator-cli.jar \
  generate \
  -i ../pkg/api/flamenco-manager.yaml \
  -g python \
  -o . \
  --skip-validate-spec \
  --package-name ${PKG_NAME} \
  --http-user-agent "Flamenco/${PKG_VERSION} (Blender add-on)" \
  -p generateSourceCodeOnly=true \
  -p projectName=Flamenco \
  -p packageVersion=${PKG_VERSION} \
