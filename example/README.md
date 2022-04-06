# Example

This example instruments a simple HTTP server and client application.

Both applications are configured to send spans to a local instance
of the OpenTelemetry Collector, which propagates them to both
Splunk Observability Cloud and to a local Jaeger instance.

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Usage

```sh
SPLUNK_ACCESS_TOKEN=<access_token> ./run.sh
```

The value for `SPLUNK_ACCESS_TOKEN` can be found
[here](https://app.signalfx.com/o11y/#/organization/current?selectedKeyValue=sf_section:accesstokens).

You can find the collected traces in:

- Console output
- Jaeger: <http://localhost:16686/search>
- Splunk Observability Cloud: <https://app.signalfx.com/#/apm?environments=YOURUSERNAME>
  (_processing takes some time_)
