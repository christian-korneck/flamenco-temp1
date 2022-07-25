---
title: Generating Code
weight: 20
---

Some code (Go, Python, JavaScript) is generated from the OpenAPI specs in
`pkg/api/flamenco-openapi.yaml`. There are also Go files generated to create
mock implementations of interfaces for unit testing purposes.

**Generated code is committed to Git**, so that after a checkout you shouldn't
need to re-run the generator to build Flamenco.

The JavaScript & Python generator is made in Java, so it requires a JRE/JDK to
be installed. On Ubuntu Linux, `sudo apt install default-jre-headless` should be
enough.

The following files & directories are generated. Generated directories are
completely erased before regeneration, so do not add any files there manually.

- `addon/flamenco/manager/`: Python API for the Blender add-on.
- `pkg/api/*.gen.go`: Go API shared by Manager and Worker.
- `internal/**/mocks/*.gen.go`: Generated mocks for Go unit tests.
- `web/app/src/manager-api/`: JavaScript API for the web front-end.
