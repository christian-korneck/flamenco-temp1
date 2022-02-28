# Flamenco 3 Blender add-on

## Setting up development environment

```
~/workspace/blender-git/build_linux/bin/3.1/python/bin/python3.9 -m venv --upgrade-deps venv
. ./venv/bin/activate
pip install poetry
poetry install
```

## Generating the OpenAPI client

Start Flamenco Manager, then run:

```
openapi-python-client generate --url http://localhost:8080/api/openapi3.json
```
