# OpenTelemetry v2

These samples use the
[OpenTelemetry v2 plugin](https://pkg.go.dev/go.temporal.io/sdk/contrib/opentelemetry-v2@v0.1.0)
for replay-safe tracing and metrics.

## Jaeger

[Jaeger](https://www.jaegertracing.io/docs/) is an open-source trace viewer. It
shows a timeline of how an operation moved through your code. These samples
export traces with OpenTelemetry and view them in Jaeger. See
[how Jaeger relates to OpenTelemetry](https://www.jaegertracing.io/docs/latest/#relationship-with-opentelemetry).

## Samples

- [Automatic instrumentation](automatic-instrumentation)
- [Workflow-to-Activity propagation](workflow-activity-propagation)
- [Client-to-Update propagation](client-update-propagation)

Shared setup lives in [`setup.go`](setup.go)
