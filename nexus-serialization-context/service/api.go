package service

const (
	Name                         = "per-endpoint-encryption-service"
	SyncOperationName            = "echo"
	AsyncOperationName           = "echo-async"
	CallerTaskQueue              = "nexus-serialization-caller"
	CompressedEncryptedTaskQueue = "nexus-serialization-compressed-encrypted-handler"
	EncryptedTaskQueue           = "nexus-serialization-encrypted-handler"
	CompressedEncrypted1Endpoint = "nexus-serialization-compressed-encrypted-1"
	CompressedEncrypted2Endpoint = "nexus-serialization-compressed-encrypted-2"
	Encrypted1Endpoint           = "nexus-serialization-encrypted-1"
	Encrypted2Endpoint           = "nexus-serialization-encrypted-2"
)

type EchoInput struct{ Message string }
type EchoOutput struct{ Message string }

// EndpointResults keeps each endpoint's synchronous and asynchronous replies together.
type EndpointResults struct {
	Sync  string
	Async string
}

type WorkflowOutput struct {
	CompressedEncrypted1 EndpointResults
	CompressedEncrypted2 EndpointResults
	Encrypted1           EndpointResults
	Encrypted2           EndpointResults
}
