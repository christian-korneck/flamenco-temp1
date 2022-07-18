---
title: Flamenco API
weight: 20
---

Flamenco Manager can be controlled via its API. This API is defined via an
[OpenAPI 3][OAPI] specification. This API can be explored via the "API" link in
the top-right corner of Flamenco's web interface. The definition itself can be
found in [pkg/api/flamenco-openapi.yaml][OAPI-YAML] in the source code.

[OAPI]: https://swagger.io/specification/
[OAPI-YAML]: https://developer.blender.org/diffusion/F/browse/main/pkg/api/flamenco-openapi.yaml

## Using the API

Flamenco API clients are generated for Go, Python, and JavaScript. These are
used by respectively the Worker, the Blender add-on, and the web frontend.

This section will explain how to translate an *OpenAPI operation* to code that
calls that operation. The examples will all use the YAML snippet below. You do
not have to understand this in detail; this guide will explain how to translate
what's written here to working code.

```yaml
paths:
  /api/v3/version:
    summary: Clients can use this to check this is actually a Flamenco server.
    get:
      summary: Get the Flamenco version of this Manager
      operationId: getVersion
      tags: [meta]
      responses:
        "200":
          description: normal response
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/FlamencoVersion"

components:
  schemas:
    FlamencoVersion:
      type: object
      properties:
        "version": { type: string }
        "name": { type: string }
      required: [version, name]
```


### JavaScript

The JavaScript code for the web-application will look something like this:

```JavaScript
import { apiClient } from '@/stores/api-query-count';
import { MetaApi } from "@/manager-api";

const metaAPI = new MetaApi(apiClient);
metaAPI.getVersion()
  .then((version) => {
    this.flamencoName = version.name;
    this.flamencoVersion = version.version;
  })
  .catch((error) => {
    console.log("Error getting the Flamenco version:", error);
  })
```

This follows a few standard steps:

1. **Import the `apiClient`**. For the web-app, this actually is a little wrapper
   for the actual API client. The wrapper takes care of showing a "loading"
   spinner in the top-right corner when API calls take a long time to return a
   response.
2. **Import the API class** that contains this operation. This always has the form
   `{tag}Api`, where `{tag}` comes from the `tag: ...` in the OpenAPI YAML
   file. In our case, the tag is `meta`, so the code imports `MetaApi`.
3. **Construct the API object** with `new MetaApi(apiClient)`. You only have to
   do this once; after that you can call as many functions on it as you want.
4. **Call** the function. The function name comes from the `operationId` in the YAML
   file,
5. **Handle** the succesful return (`.then(...)`) and any errors (`.catch(...)`).

All API function calls, like the `metaAPI.getVersion()` function above,
immediately return a [promise][promise]. This means that any code after the API
call will be immediately executed; it will not wait for the API to respond.

[promise]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Using_promises

The object passed to the `.then(...)` callback also follows the OpenAPI
specification. It has `name` and `version` properties, because those were
declared in its *schema*.

### Python

The Python code for this call will look something like this:

```Python
from flamenco.manager import ApiClient, Configuration
from flamenco.manager.apis import MetaApi
from flamenco.manager.models import FlamencoVersion

configuration = Configuration(host="http://localhost:8080")
api_client = ApiClient(configuration)

meta_api = MetaApi(api_client)
version: FlamencoVersion = meta_api.get_version()
print(f"Found {version.name} version {version.version}")
```

This follows a few standard steps:

1. **Create an API client**, where the URL is the base URL of the Flamenco Manager
   to communicate with.
2. **Construct the API object** with `MetaApi(api_client)`. This always has the
   form `{tag}Api`, where `{tag}` comes from the `tag: ...` in the OpenAPI YAML
   file. In our case, the tag is `meta`, so the code imports `MetaApi`. You only
   have to do this once; after that you can call as many functions on it as you
   want.
3. **Call** the function. The function name comes from the `operationId` in the
   YAML file, converted to [snake case][snake].
4. **Handle** the returned value. The call is synchronous, and thus will block
   until the Manager has responded.

API calls can raise an `flamenco.manager.ApiException` exception if there was
some API-level error, like validation errors from Flamenco Manager. They can
also raise any exception from `urllib3.exceptions` when there are errors in the
HTTP or TCP/IP stack, such as the Manager being unreachable.

[snake]: https://en.wikipedia.org/wiki/Snake_case

### Go

The Go code for this call will look something like this:

```Go
package main

import (
	"context"
	"fmt"

	"git.blender.org/flamenco/pkg/api"
)

func main() {
	client, err := api.NewClientWithResponses("http://localhost:8080")
  if err != nil {
    panic(err)
  }

	ctx := context.Background()
	resp, err := client.GetVersionWithResponse(ctx)

	switch {
	case err != nil: // This will be generic errors, like the Manager not being reachable.
		panic(err)
	case resp.JSON200 == nil: // resp.JSON200 is a pointer to the "200" response from the YAML.
		// If we did not get a normal response, `resp.HTTPResponse` can be inspected.
		panic(resp.HTTPResponse.Status)
	}

	version := resp.JSON200
	fmt.Printf("Found %s version %s\n", version.Name, version.Version)
}
```

This follows a few standard steps:

1. **Create an API client**, where the URL is the base URL of the Flamenco Manager
   to communicate with.
2. **Call** the function. The function name comes from the `operationId` in the
   YAML file, starting with a capital letter to make it accessible for other
   packages, and suffixed with `WithResponse`.
3. **Handle** the returned value. The call is synchronous, and thus will block
   until the Manager has responded.

Contrary to the other API clients, the Go client does not use the tags to group
functions.
