#!/bin/bash
set -euo pipefail

FUNCTION_NAME="${1:?Usage: deploy-lambda.sh <function-name>}"

GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ./worker
zip function.zip bootstrap client.pem client.key temporal.toml otel-collector-config.yaml
aws lambda update-function-code --function-name "$FUNCTION_NAME" --zip-file fileb://function.zip
