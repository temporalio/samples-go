# Codec Server: protocol

# Summary

This document outlines the HTTP protocol for codec servers. This functionality allows users to deploy a codec centrally rather than the previous architecture that required a `tctl` plugin on developer workstations. This makes it easier to secure access to any required encryption keys and simplifies the developer experience.

# Protocol

The codec HTTP protocol specifies two endpoints, one for encoding a Payloads object and one for decoding.

Implementations MUST:

1. Send and receive [Payloads](https://github.com/temporalio/api/blob/e82978c745a07fb8820348ad77b1d02e226d182e/temporal/api/common/v1/message.proto#L46) protobuf as JSON per [https://developers.google.com/protocol-buffers/docs/proto3#json](https://developers.google.com/protocol-buffers/docs/proto3#json).
Implementations should not rely on the standard JSON encoding of objects in their language but must use Protobuf specific JSON encoders. Libraries are available to handle this for most languages.
2. Only check the final part of the incoming URL to determine if the request is for /encode or /decode.
This makes deployment more flexible by allowing the endpoints to be mounted at any depth in a URL hierarchy, for example to allow encoders for different namespaces to be served from the same hostname.

Implementations MAY:

1. Support codec for different namespaces under different URLs.
2. Read the `X-Namespace` header sent to the /encode or /decode endpoints as an alternative to differentiating namespaces based on URL. The current `tctl` and Temporal Web UI codec client code will set `X-Namespace` appropriately for each request.

In the endpoint sequence diagrams below we are using `tctl` as an example of the client side, but Temporal Web and all other consumers will follow the same protocol.

## Encode

```mermaid
sequenceDiagram;
	participant tctl
	participant Server as Codec Server

	tctl->>Server: HTTP POST /encode
	Note right of tctl: Content-Type: application/json
	Note right of tctl: Body: Payloads protobuf as JSON
	alt invalid JSON
		Server-->>tctl: HTTP 400 BadRequest
    else decoder error
		Server-->>tctl: HTTP 400 BadRequest
    else
		Server-->>tctl: HTTP 200 OK
		Note left of Server: Content-Type: application/json
		Note left of Server: Body: Encoded Payloads protobuf as JSON
	end
```

## Decode

```mermaid
sequenceDiagram;
	participant tctl
	participant Server as Codec Server

	tctl->>Server: HTTP POST /decode
	Note right of tctl: Content-Type: application/json
	Note right of tctl: Body: Payloads protobuf as JSON
	alt invalid JSON
		Server-->>tctl: HTTP 400 BadRequest
  else decoder error
		Server-->>tctl: HTTP 400 BadRequest
  else
		Server-->>tctl: HTTP 200 OK
		Note left of Server: Content-Type: application/json
		Note left of Server: Body: Decoded Payloads protobuf as JSON
	end

```
