# Example

This example instruments a simple HTTP server-client application.

The application is configured to send spans and metrics
to a local instance of the OpenTelemetry Collector,
which propagates them to Splunk Observability Cloud.

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Usage

### OpenTelmemtry Collector

Run the OpenTelemetry Collector and Jaeger instance:

```sh
SPLUNK_ACCESS_TOKEN=<access_token> docker compose up -d
```

The value for `SPLUNK_ACCESS_TOKEN` can be found
[here](https://app.signalfx.com/o11y/#/organization/current?selectedKeyValue=sf_section:accesstokens).
Reference: [docs](https://docs.splunk.com/Observability/admin/authentication-tokens/api-access-tokens.html#admin-api-access-tokens).

Run the instrumented application:

```sh
export OTEL_SERVICE_NAME="splunk-otel-go-example"
export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=$(whoami)"
go run .
```

You can find the collected telemetry in:

- OpenTelemetry Collector output
- Jaeger: <http://localhost:16686/search>
- Prometheus scrape handler: <http://localhost:8889/metrics>
- Splunk Observability Cloud: <https://app.signalfx.com/#/apm?environments=YOURUSERNAME>
  > Note: Processing might take some time.

Cleanup:

```sh
docker compose down
```

### Splunk Distribution of the OpenTelemetry Collector

Run the Splunk Distribution of the OpenTelemetry Collector instance:

```sh
SPLUNK_ACCESS_TOKEN=<access_token> docker compose -f docker-compose-splunk.yaml up -d
```

The value for `SPLUNK_ACCESS_TOKEN` can be found
[here](https://app.signalfx.com/o11y/#/organization/current?selectedKeyValue=sf_section:accesstokens).
Reference: [docs](https://docs.splunk.com/Observability/admin/authentication-tokens/api-access-tokens.html#admin-api-access-tokens).

Run the instrumented application:

```sh
export OTEL_SERVICE_NAME="splunk-otel-go-example"
export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=$(whoami)"
go run .
```

You can find the collected telemetry in:

- Splunk Observability Cloud: <https://app.signalfx.com/#/apm?environments=YOURUSERNAME>
  > Note: Processing might take some time.

Cleanup:

```sh
docker compose -f docker-compose-splunk.yaml down
```

### Direct to Splunk Observability Cloud

Run the instrumented application:

```sh
export OTEL_SERVICE_NAME="splunk-otel-go-example"
export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=$(whoami)"
SPLUNK_REALM=<realm> SPLUNK_ACCESS_TOKEN=<access_token> go run .
```

You can find the collected telemetry in:

- Splunk Observability Cloud: <https://app.signalfx.com/#/apm?environments=YOURUSERNAME>
  > Note: Processing might take some time.

### FIPS mode - Linux

> [!NOTE]  
> As BoringSSL is FIPS 140-2 certified, an application built using `GOEXPERIMENT=boringcrypto`
> is more likely to be FIPS 140-2 compliant.
> Yet Google does not provide any liability about the suitability of this code
> in relation to the FIPS 140-2 standard.
> More information can be found [here](https://go.dev/src/crypto/internal/boring/README).

Run the instrumented applications using
[`boringcrypto`](https://github.com/microsoft/go/blob/microsoft/main/eng/doc/fips/README.md#go-fips-compliance).
For example:

```sh
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go run .
```
