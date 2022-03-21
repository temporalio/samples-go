package pso

import (
	"bytes"
	"encoding/json"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

// jsonDataConverter implements converter.DataConverter using JSON for Swarm and Particle
// WARGNING: Make sure all struct members are public (Capital letter) otherwise serialization does not work!
// TODO: consider storing blobs in external DB or S3
type jsonDataConverter struct {
}

// NewJSONDataConverter creates a json data converter
func NewJSONDataConverter() converter.DataConverter {
	return &jsonDataConverter{}
}

// Json data converter implementation

func (dc *jsonDataConverter) ToPayloads(value ...interface{}) (*commonpb.Payloads, error) {
	payloads := &commonpb.Payloads{}
	for _, obj := range value {
		payload, err := dc.ToPayload(obj)
		if err != nil {
			return nil, err
		}

		payloads.Payloads = append(payloads.Payloads, payload)
	}

	return payloads, nil
	// TODO: store payloads in DB/S3 and return converter key
	// return key, nil
}

func (dc *jsonDataConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	var err error
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	switch t := value.(type) {
	case Swarm:
		err = enc.Encode(*t.Settings)
		if err == nil {
			err = enc.Encode(*t.Gbest)
			if err == nil {
				if t.Settings.Size > 0 {
					for _, particle := range t.Particles {
						if particle == nil {
							particle = new(Particle)
						}
						err = enc.Encode(*particle)
					}
				}
			}
		}
	case WorkflowResult:
		err = enc.Encode(t.Msg)
		if err == nil {
			err = enc.Encode(t.Success)
		}
	default:
		err = enc.Encode(value)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"unable to encode argument: %T, with error: %w", value, err)
	}

	payload := &commonpb.Payload{
		Metadata: map[string][]byte{
			"encoding": []byte("raw"),
		},
		Data: buf.Bytes(),
	}

	return payload, nil
}

func (dc *jsonDataConverter) FromPayloads(payloads *commonpb.Payloads, valuePtrs ...interface{}) error {
	if payloads == nil {
		return nil
	}
	// TODO: convert payloads into key in DB/S3 and retrieve actual payloads from DB/S3
	for i, payload := range payloads.Payloads {
		err := dc.FromPayload(payload, valuePtrs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (dc *jsonDataConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	var err error
	dec := json.NewDecoder(bytes.NewBuffer(payload.GetData()))
	switch t := valuePtr.(type) {
	case *Swarm:
		t.Settings = new(SwarmSettings)
		_ = dec.Decode(t.Settings)
		t.Settings.function = FunctionFactory(t.Settings.FunctionName)
		t.Gbest = NewPosition(t.Settings.function.dim)
		err = dec.Decode(t.Gbest)
		t.Particles = make([]*Particle, t.Settings.Size)
		for index := 0; index < t.Settings.Size; index++ {
			t.Particles[index] = new(Particle)
			err = dec.Decode(t.Particles[index])
		}
	case *WorkflowResult:
		err = dec.Decode(&t.Msg)
		if err == nil {
			err = dec.Decode(&t.Success)
		}
	default:
		err = dec.Decode(valuePtr)
	}
	if err != nil {
		return fmt.Errorf(
			"unable to decode argument: %T, with error: %v", valuePtr, err)
	}
	return nil
}

func (dc *jsonDataConverter) ToString(_ *commonpb.Payload) string {
	return "implement me"
}

func (dc *jsonDataConverter) ToStrings(_ *commonpb.Payloads) []string {
	return []string{"implement me"}
}
