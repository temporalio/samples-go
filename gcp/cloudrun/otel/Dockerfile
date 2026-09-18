FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/cloud-run-worker ./gcp/cloudrun/otel/worker

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /out/cloud-run-worker /cloud-run-worker
ENTRYPOINT ["/cloud-run-worker"]
