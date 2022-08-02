---
title: Building Flamenco
weight: 10
---

For the steps towards your first build, see [Getting Started][start].

[start]: {{< relref "../getting-started ">}}

These `make` targets are available:

| Target                  | Description                                                                                                                                                                                |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `application`           | Builds Flamenco Manager, Worker, and the development version of the webapp.                                                                                                                |
| `flamenco-manager`      | Builds just Flamenco Manager. This includes packing the webapp and the Blender add-on into the executable.                                                                                 |
| `flamenco-worker`       | Builds just Flamenco Worker.                                                                                                                                                               |
| `flamenco-manager_race` | Builds Flamenco Manager with the [data race detector][race] enabled. As this is for development only, this does not include packing the webapp and the Blender add-on into the executable. |
| `flamenco-worker_race`  | Builds Flamenco Worker with the [data race detector][race] enabled.                                                                                                                        |
| `webapp`                | Installs the webapp dependencies, so that the development server can be run with `yarn --cwd web/app run dev --host`                                                                       |
| `webapp-static`         | Builds the webapp so that it can be served as static files by Flamenco Manager.                                                                                                            |
| `addon-packer`          | Builds the addon packer. This is a little Go tool that creates the Blender add-on ZIP file. Typically this target isn't used directly; the other Makefile targets depend on it.            |
| `generate`              | Generate the Go, Python, and JavaScript code.                                                                                                                                              |
| `generate-go`           | Generate the Go code, which includes OpenAPI code, as well as mocks for the unit tests.                                                                                                    |
| `generate-py`           | Generate the Python code, containing the OpenAPI client code for the Blender add-on.                                                                                                       |
| `generate-js`           | Generate the JavaScript code, containing the OpenAPI client code for the web interface.                                                                                                    |
| `test`                  | Run the unit tests.                                                                                                                                                                        |
| `clean`                 | Remove build-time files.                                                                                                                                                                   |
| `version`               | Print some version numbers, mostly for debugging the Makefile itself.                                                                                                                      |
| `list-embedded`         | List the files embedded into the `flamenco-manager` executable.                                                                                                                            |

[race]: https://go.dev/doc/articles/race_detector
