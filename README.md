# Flamenco PoC

This repository contains a proof of concept of a next-generation Flamenco implementation.

## Building

1. Install [Go 1.18 or newer](https://go.dev/) and Node 16 (see below)
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
sudo -u postgres createdb -O flamenco -E utf8 flamenco
sudo -u postgres createdb -O flamenco -E utf8 flamenco-test
echo "alter schema public owner to flamenco;" | sudo -u postgres psql flamenco-test
```

### Windows

On Windows, add `C:\Program Files\PostgreSQL\14\bin` to your `PATH` environment variable.
Replace `14` with the version of PostgreSQL you're using. Then run:


```
createuser -U postgres -D -P flamenco  # give it the password 'flamenco'
createdb -U postgres -O flamenco -E utf8 flamenco
createdb -U postgres -O flamenco -E utf8 flamenco-test
psql -c "alter schema public owner to flamenco" flamenco-test postgres
```

When it asks "Enter password for new role:", give the password "flamenco"
When it asks "Password:", give the password for the postgres admin user (you chose this during installation of PostgreSQL).

If you're like me, and you use Git Bash, prefix the commands with `winpty`.
