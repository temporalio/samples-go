#!/bin/bash
set -euo pipefail

# Creates the IAM role that allows Temporal Cloud to invoke your Lambda function.
# You can find the External ID in your Temporal Cloud namespace settings.

STACK_NAME="${1:?Usage: mk-iam-role.sh <stack-name> <external-id> <lambda-arn>}"
EXTERNAL_ID="${2:?Usage: mk-iam-role.sh <stack-name> <external-id> <lambda-arn>}"
LAMBDA_ARN="${3:?Usage: mk-iam-role.sh <stack-name> <external-id> <lambda-arn>}"

aws cloudformation create-stack \
  --stack-name "$STACK_NAME" \
  --template-body file://iam-role-for-temporal-lambda-invoke.yaml \
  --parameters \
    ParameterKey=AssumeRoleExternalId,ParameterValue="$EXTERNAL_ID" \
    ParameterKey=LambdaFunctionARN,ParameterValue="$LAMBDA_ARN" \
  --capabilities CAPABILITY_NAMED_IAM
