// Package runtime — Codec factory.
//
// A Codec is the read/write pair.  Different serialisation formats
// (JSON, YAML, protobuf) plug in here; AxiomNizam only needs JSON and
// YAML for now, so those are the only two provided.  Callers obtain a
// Codec from a CodecFactory bound to a Scheme.
package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Encoder serialises an Object to bytes.
type Encoder interface {
	Encode(obj Object, w io.Writer) error
}

// Decoder parses bytes into an Object.  The into parameter is
// optional — when nil, the decoder uses the Scheme to dispatch.
type Decoder interface {
	Decode(data []byte, into Object) (Object, error)
}

// Codec bundles both halves.
type Codec interface {
	Encoder
	Decoder
}

// CodecFactory constructs Codecs bound to a Scheme.
type CodecFactory struct {
	scheme *Scheme
}

// NewCodecFactory returns a factory for scheme.
func NewCodecFactory(scheme *Scheme) CodecFactory {
	return CodecFactory{scheme: scheme}
}

// UniversalDeserializer returns a decoder that can handle any JSON
// payload whose TypeMeta is registered.
func (f CodecFactory) UniversalDeserializer() Decoder {
	return &jsonCodec{scheme: f.scheme}
}

// LegacyCodec returns a JSON Codec.  "Legacy" is a carry-over from
// upstream naming — AxiomNizam has no protobuf variant yet.
func (f CodecFactory) LegacyCodec() Codec {
	return &jsonCodec{scheme: f.scheme}
}

// jsonCodec is the concrete implementation.
type jsonCodec struct {
	scheme *Scheme
}

// Encode stamps TypeMeta and writes compact JSON to w.
func (c *jsonCodec) Encode(obj Object, w io.Writer) error {
	kinds, err := c.scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	if len(kinds) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(kinds[0])
	}
	buf, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

// Decode routes based on TypeMeta when into is nil; otherwise it
// unmarshals straight into into and returns it.
func (c *jsonCodec) Decode(data []byte, into Object) (Object, error) {
	if into != nil {
		if err := json.Unmarshal(data, into); err != nil {
			return nil, err
		}
		return into, nil
	}
	return c.scheme.Decode(data)
}

// EncodeToBytes is a convenience helper for the common case.
func EncodeToBytes(c Codec, obj Object) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.Encode(obj, &buf); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	return buf.Bytes(), nil
}
