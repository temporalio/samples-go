package pso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	commonpb "go.temporal.io/temporal-proto/common"
	"go.temporal.io/temporal/encoded"
)

// jsonDataConverter implements encoded.DataConverter using JSON for Swarm and Particle
// WARGNING: Make sure all struct members are public (Capital letter) otherwise serialization does not work!
// TODO: consider storing blobs in external DB or S3
type jsonDataConverter struct {
}

// NewJSONDataConverter creates a json data converter
func NewJSONDataConverter() encoded.DataConverter {
	return &jsonDataConverter{}
}

// Json data converter implementation

func (dc *jsonDataConverter) ToData(value ...interface{}) (*commonpb.Payloads, error) {
	payloads := &commonpb.Payloads{}
	for i, obj := range value {
		var err error
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)

		switch t := obj.(type) {
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
			err = enc.Encode(obj)
		}
		if err != nil {
			return nil, fmt.Errorf(
				"unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}

		payloads.Payloads = append(payloads.Payloads, &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte("raw"),
			},
			Data: buf.Bytes(),
		})
	}

	return payloads, nil
	// TODO: store payloads in DB/S3 and return encoded key
	// return key, nil
}

func (dc *jsonDataConverter) FromData(payloads *commonpb.Payloads, valuePtr ...interface{}) error {
	// TODO: convert payloads into key in DB/S3 and retrieve actual payloads from DB/S3
	for i, payload := range payloads.Payloads {
		var err error
		obj := valuePtr[i]
		dec := json.NewDecoder(bytes.NewBuffer(payload.GetData()))
		switch t := obj.(type) {
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
			err = dec.Decode(obj)
		}
		if err != nil {
			return fmt.Errorf(
				"unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}
