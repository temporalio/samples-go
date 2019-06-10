package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"

	"go.uber.org/cadence/encoded"
)

// gobDataConverter implements encoded.DataConverter using gob for Swarm and Particle
// WARGNING: Make sure all struct members are public (Capital letter) otherwise serialization does not work!
// TODO: consider storing blobs in external DB or S3
type gobDataConverter struct {
}

func (dc *gobDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	var err error
	for i, obj := range value {
		switch t := obj.(type) {
		case Swarm:
			err = enc.Encode(*t.Settings)
			err = enc.Encode(*t.Gbest)
			if t.Settings.Size > 0 {
				for _, particle := range t.Particles {
					if particle == nil {
						particle = new(Particle)
					}
					err = enc.Encode(*particle)
				}
			}
		default:
			err = enc.Encode(obj)
		}
		if err != nil {
			return nil, fmt.Errorf(
				"unable to encode argument: %d, %v, with gob error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return buf.Bytes(), nil
	// TODO: store buf.Bytes() in DB/S3 and get key
	// return key, nil
}

func (dc *gobDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	// TODO: convert input into key in DB/S3 and retrieve bytes
	//dec := gob.NewDecoder(bytes)
	dec := gob.NewDecoder(bytes.NewBuffer(input))
	var err error
	for i, obj := range valuePtr {
		switch t := obj.(type) {
		case *Swarm:
			t.Settings = new(SwarmSettings)
			err = dec.Decode(t.Settings)
			t.Settings.Function = FunctionFactory(t.Settings.FunctionName)
			t.Gbest = NewPosition(t.Settings.Function.dim)
			err = dec.Decode(t.Gbest)
			t.Particles = make([]*Particle, t.Settings.Size)
			for index := 0; index < t.Settings.Size; index++ {
				t.Particles[index] = new(Particle)
				err = dec.Decode(t.Particles[index])
			}
		default:
			err = dec.Decode(obj)
		}
		if err != nil {
			return fmt.Errorf(
				"unable to decode argument: %d, %v, with gob error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

// NewGobDataConverter creates a gob data converter
func NewGobDataConverter() encoded.DataConverter {
	return &gobDataConverter{}
}
