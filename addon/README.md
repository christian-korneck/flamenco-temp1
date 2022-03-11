# Flamenco 3 Blender add-on

## Setting up development environment

```
~/workspace/blender-git/build_linux/bin/3.1/python/bin/python3.9 -m venv --upgrade-deps venv
. ./venv/bin/activate
pip install poetry
poetry install
```

## Generating the OpenAPI client

1. Make sure Java is installed (so `java --version` shows something sensible).
2. In the root directory of the repository, run `make generate-py`
