# Container build works once cloudrun/id is released and the go.mod replace is removed.
FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/cloud-run-id ./gcp/cloudrun/id/worker

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /out/cloud-run-id /cloud-run-id
ENTRYPOINT ["/cloud-run-id"]
