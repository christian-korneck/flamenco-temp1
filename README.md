# Flamenco PoC

This repository contains a proof of concept of a next-generation Flamenco implementation.

## Building

1. Install [Go 1.17 or newer](https://go.dev/).
2. Set the environment variable `GOPATH` to where you want Go to put its packages. Defaults to `$HOME/go` if not set. Run `go env GOPATH` if you're not sure.
3. Ensure `$GOPATH/bin` is included in your `$PATH` environment variable.
4. Run the following commands:

```
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen
make
```

You should now have two executables: `flamenco-manager-poc` and `flamenco-worker-poc`.

## Swagger UI

Flamenco Manager has a SwaggerUI interface at http://localhost:8080/api/swagger-ui/
