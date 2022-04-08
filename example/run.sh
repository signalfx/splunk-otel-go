#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=$(whoami)"
docker-compose -f "${DIR}/docker-compose.yaml" up
