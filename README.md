# Flamenco 3

This repository contains the sources for Flamenco 3. The Manager, Worker, and
Blender add-on sources are all combined in this one repository.


## Building

1. Install [Go 1.18 or newer](https://go.dev/), Java (just a JRE is enough), and Node 16 (see below)
2. Set the environment variable `GOPATH` to where you want Go to put its packages. Defaults to `$HOME/go` if not set. Run `go env GOPATH` if you're not sure.
3. Ensure `$GOPATH/bin` is included in your `$PATH` environment variable.
4. Magically build the web frontend (still under development, no concrete steps documentable quite yet)
5. Run `make with-deps` to install build-time dependencies and build the application. Subsequent builds can just run `make` without arguments.

You should now have two executables: `flamenco-manager` and `flamenco-worker`.


## Node / Web UI

The web UI is built with Vue, Bootstrap, and Socket.IO for communication with the backend.

NodeJS is used to collect all of those and build the frontend files. It's recommended to install Node v16 via Snap:

```
sudo snap install node --classic --channel=16
```

This also gives you the Yarn package manager, which can be used to install web dependencies and build the frontend files.


## Swagger UI

Flamenco Manager has a SwaggerUI interface at http://localhost:8080/api/swagger-ui/


## Database

Flamenco Manager and Worker use SQLite as database, and Gorm as
object-relational mapper.

Since SQLite has limited support for altering table schemas, migration requires
copying old data to a temporary table with the new schema, then swap out the
tables. Because of this, avoid `NOT NULL` columns, as they will be problematic
in this process.


## License

Flamenco is licensed under the GPLv3+ license.
