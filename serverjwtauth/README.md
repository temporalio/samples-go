# Temporal JWT Authorization

The Temporal server may be configured to authorize based on JWT sent as a header. The server communicates with an HTTP
endpoint that hosts the public keys used to sign tokens (in JWK format, ref
[RFC 7517](https://datatracker.ietf.org/doc/html/rfc7517)). Once the key's signature has been validated, the
`permissions` claim in the token is used to determine what the caller can do. See
[Temporal security documentation](https://docs.temporal.io/docs/server/security) for more details.

This example creates a ECDSA key pair, serves the public key for use by the server, and creates keys as needed for
using `tctl`, running workers, and starting workflows.

Note: Since we are not authorizing the UI in this sample, the server UI will not work with this restricted access.
[Configuration of SSO for the Temporal Web UI](https://github.com/temporalio/web#configuring-authentication-optional) is
outside the scope of this example.

## Running

### Starting the server

First, the public keys must be served for reading by the server. Run the following in a separate terminal to generate a
private key and serve its public key:

    go run ./serverjwtauth/key gen-and-serve

This will output a line like:

    Started JWKS server. Endpoint: http://[::]:61884/jwks.json. Ctrl+C to exit.

This creates a file at `key.priv.pem` with the private key. The port listed on each run will likely be different, so
replace `61884` with the output port in all commands henceforth.

With this running we can start the server. See the primary README about starting the server.
[docker-compose](https://github.com/temporalio/docker-compose) is suggested for this sample.

Before starting the server you must configure authorization. If using the docker image, the following environment
variables should be set:

* `TEMPORAL_JWT_KEY_SOURCE1=http://host.docker.internal:61884/jwks.json`
* `TEMPORAL_AUTH_AUTHORIZER=default`
* `TEMPORAL_AUTH_CLAIM_MAPPER=default`

If using the docker-compose local install, these can be set in `docker-compose.yml`. If using a standalone server
locally, the config is:

```yaml
global:
  authorization:
    jwtKeyProvider:
      keySourceURIs:
      - http://127.0.0.1:61884/jwks.json
      refreshInterval: 1m
    authorizer: default
    claimMapper: default
```

With these set, start the server in the background via whatever install method chosen.

### Using `tctl` and registering the `default` namespace

When starting the server using the docker container, the following excerpt will appear in the logs:

    "msg":"uncategorized error","operation":"RegisterNamespace","wf-namespace":"default","error":"Request unauthorized."

This is because by default the docker container attempts to register the `default` namespace, but cannot because we have
restricted access to the server to only authorized uses.

To create this manually, we will use `tctl`. According to
[the tctl docs](https://docs.temporal.io/docs/devtools/tctl/#securing-tctl) we can use the `tctl-authorization-plugin`
binary. If using docker, this binary is included with tctl otherwise, build it and put it on the `PATH`.

We must set the `TEMPORAL_CLI_AUTHORIZATION_TOKEN` environment variable with an authorization header value that includes
the JWT key. The following command generates a 1 hour key with `admin` permissions on the `system` namespace:

    go run ./serverjwtauth/key tctl-system-token

This will output something like:

    TEMPORAL_CLI_AUTHORIZATION_TOKEN=Bearer abcde...

See [this documentation](https://docs.temporal.io/docs/devtools/tctl/#run-the-cli) on how to run `tctl` with environment
variables. If using the docker approach from bash, an argument could be added like:

    --env "$(go run ./serverjwtauth/key tctl-system-token)"

Or if using `tctl` standalone from bash, an `export` for the environment can be added:

    export "$(go run ./serverjwtauth/key tctl-system-token)"

Once `tctl` is set to take an environment variable, it can be run to create the `default` namespace:

    tctl --headers_provider_plugin tctl-authorization-plugin --ns default namespace register -rd 1

We pass `--headers_provider_plugin tctl-authorization-plugin` so the environment variable is used for authentication. A
more advanced use case could very easily create a new plugin binary for the header provider in [keyutil.go](keyutil.go)
but that is beyond the scope of this sample.

### Running the worker and starting the workflow

Now with the server running with authorization enabled and the namespace present, the worker can be run in the
background or a separate terminal:

    go run ./serverjwtauth/worker

With the worker started, we can run a workflow:

    go run ./serverjwtauth/starter

This will output:

    Workflow result: Hello Temporal!

Both the `worker` and `starter` are configured to dynamically create/rotate JWTs based on the private key. The JWT they
use is configured to only have permissions for read/write on the `default` namespace.
