# Example

This example instruments a simple HTTP server and client application.

Both applications are configured to send spans to a local instance
of the OpenTelemetry Collector, which propagates them to both
Splunk Observability Cloud and to a local Jaeger instance.

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Usage

Running:

```sh
SPLUNK_ACCESS_TOKEN=<access_token> ./run.sh
```

The value for `SPLUNK_ACCESS_TOKEN` can be found
[here](https://app.signalfx.com/o11y/#/organization/current?selectedKeyValue=sf_section:accesstokens).
Reference: [docs](https://docs.splunk.com/Observability/admin/authentication-tokens/api-access-tokens.html#admin-api-access-tokens).

You can find the collected traces in:

- Console output
- Jaeger: <http://localhost:16686/search>
- Splunk Observability Cloud: <https://app.signalfx.com/#/apm?environments=YOURUSERNAME>
  > Note: Processing might take some time.

Cleanup:

```sh
./clean.sh
```
