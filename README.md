# Flamenco 3

This repository contains the sources for Flamenco 3. The Manager, Worker, and
Blender add-on sources are all combined in this one repository.


## Building

1. Install [Go 1.18 or newer](https://go.dev/), and Node 16 (see "Node / Web UI" below).
2. Optional: set the environment variable `GOPATH` to where you want Go to put its packages.
3. Ensure `$GOPATH/bin` is included in your `$PATH` environment variable. `$GOPATH` defaults to `$HOME/go` if not set. Run `go env GOPATH` if you're not sure.
4. Set up the web frontend for development (see "Node / Web UI" below).
5. Run `make with-deps` to install build-time dependencies and build the application. Subsequent builds can just run `make` without arguments.

You should now have two executables: `flamenco-manager` and `flamenco-worker`.
Both can be run with the `-help` CLI argument to see the available options.

To rebuild only the Manager or Worker, run `make flamenco-manager` or `make flamenco-worker`.


## Node / Web UI

The web UI is built with Vue, Bootstrap, and Socket.IO for communication with the backend. NodeJS/NPM is used to collect all of those and build the frontend files. It's recommended to install Node v16 via Snap:

```
sudo snap install node --classic --channel=16
```

This also gives you the Yarn package manager, which can be used to install web dependencies and build the frontend files.

```
cd web/app
yarn install
```

Then run the frontend development server with:
```
yarn run dev --host
```

The `--host` parameter is optional but recommended. The downside is that it
exposes the devserver to others on the network. The upside is that it makes it
easier to detect configuration issues. The generated OpenAPI client defaults to
using `localhost`, and if you're not testing on `localhost` this stands out
more.


## Generating the OpenAPI/Swagger API

Some code is generated from the OpenAPI specs in
`pkg/api/flamenco-manager.yaml`. The generated code is committed to Git, so that
after a checkout you shouldn't need to re-run the generator to build Flamenco.

After changing `pkg/api/flamenco-manager.yaml`, run `make generate` to generate
the code, then commit to Git.

The JavaScript and Python generator is made in Java, so it requires a JRE/JDK to
be installed. On Ubuntu Linux, `sudo apt install default-jre-headless` should be
enough.


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
