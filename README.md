# Flamenco PoC

This repository contains a proof of concept of a next-generation Flamenco implementation.

## Building

1. Install [Go 1.17 or newer](https://go.dev/) and Node 16 (see below)
2. Set the environment variable `GOPATH` to where you want Go to put its packages. Defaults to `$HOME/go` if not set. Run `go env GOPATH` if you're not sure.
3. Ensure `$GOPATH/bin` is included in your `$PATH` environment variable.
4. Magically build the web frontend (still under development, no concrete steps documentable quite yet)
5. Run `make with-deps` to install build-time dependencies and build the application. Subsequent builds can just run `make` without arguments.

You should now have two executables: `flamenco-manager-poc` and `flamenco-worker-poc`.


## Node / Web UI

The web UI is built with Vue, Bootstrap, and Socket.IO for communication with the backend.

NodeJS is used to collect all of those and build the frontend files. It's recommended to install Node v16 via Snap:

```
sudo snap install node --classic --channel=16
```

This also gives you the Yarn package manager, which can be used to install web dependencies and build the frontend files.

## Swagger UI

Flamenco Manager has a SwaggerUI interface at http://localhost:8080/api/swagger-ui/

## Flamenco Manager DB development machine setup.

Install PostgreSQL, then run:

```
sudo -u postgres createuser -D -P flamenco  # give it the password 'flamenco'
sudo -u postgres createdb flamenco -O flamenco -E utf8
sudo -u postgres createdb flamenco-test -O flamenco -E utf8
echo "alter schema public owner to flamenco;" | sudo -u postgres psql flamenco-test
```
