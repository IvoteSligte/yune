package oneof

// based on github.com/dhoelle/oneof
// but with many changes since the JSON V2 API has changed since its last release

import (
	"fmt"

	"github.com/go-json-experiment/json"
)

type Option interface {
	OneofKey() string
}

// UnmarshalFunc creates a [json.UnmarshalFuncV2] which will intercept
// unmarshaling behavior for values of type T.
//
// UnmarshalFunc finds the destination Go type in its option set that matches
// the JSON type discriminator, then decodes the remaining JSON according to the
// default JSON encoding of T.
func UnmarshalFunc[T any](opts []Option) *json.Unmarshalers {
	var unmarshalOpts *json.Unmarshalers

	unmarshalFunc := func(bytes []byte, ptr *T) error {

		// We expect the JSON for this type to be wrapped in a
		// way that tells us what type of T we should decode into.
		//
		// So first, decode the input into our discriminated type.
		discriminated := struct {
			Type string `json:"_type"`
		}{}
		if err := json.Unmarshal(bytes, &discriminated, json.WithUnmarshalers(unmarshalOpts)); err != nil {
			return fmt.Errorf("Failed to decode to type wrapper: %w", err)
		}
		// ...then, extract the type and use it to select a T
		// from our options
		opt, ok := discriminatorToOption(discriminated.Type, opts)
		if !ok {
			panic("Unknown discriminator value: " + discriminated.Type)
		}

		// ...then, unmarshal the remainder into the selected
		// option
		optPtr := &opt

		if err := json.Unmarshal(bytes, optPtr); err != nil {
			return fmt.Errorf("Failed to marshal value to option type %T: %w", opt, err)
		}
		*ptr = opt.(T)
		return nil
	}
	unmarshalOpts = json.UnmarshalFunc(unmarshalFunc)
	return unmarshalOpts
}

func discriminatorToOption(d string, opts []Option) (found Option, ok bool) {
	for _, opt := range opts {
		if opt.OneofKey() == d {
			found = opt
			ok = true
			return
		}
	}
	return
}
