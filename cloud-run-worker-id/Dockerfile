# Builds the Cloud Run worker image.
#
# NOTE: While the go.temporal.io/sdk/contrib/gcp/cloudrun module is unreleased, samples-go/go.mod
# carries a local `replace` to ../sdk-go-2. A container build's context is the samples repo, so it
# cannot reach that sibling directory: this Dockerfile builds as-is only once the module is
# published and the replace is removed. See the README for details.
FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/cloud-run-worker-id ./cloud-run-worker-id/worker

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /out/cloud-run-worker-id /cloud-run-worker-id
ENTRYPOINT ["/cloud-run-worker-id"]
