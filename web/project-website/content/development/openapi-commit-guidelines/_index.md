---
title: OpenAPI Commit Guidelines
weight: 30
---

Typically a change to the OpenAPI definition consists of three steps, namely
making the change to the OpenAPI file, regenerating code, and then alter
whatever manually-written code needs altering.

Each of these steps should be **committed independently**, by following these
steps:

1. Commit the changes to `pkg/api/flamenco-openapi.yaml`, prefixing the commit
   message with `OAPI:`.
2. Regenerate code with `make generate`, then commit with message
   `OAPI: Regenerate code`.
3. Commit any other code changes to complete the change.

The downside to this approach is that the second commit will likely break the
project, which is unfortunate. However, this approach does have some advantages:

- The regenerated code has the commit hash of the actual change to the OpenAPI
  definition.
- Changes to manually-written and generated code are tracked in separate
  commits. This makes them easier to comprehend by humans.
