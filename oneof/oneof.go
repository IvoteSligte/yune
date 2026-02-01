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
	// Hack:
	//
	// Our strategy for generically decoding JSON into a Go
	// type that implements T is to:
	//
	//  1. Intercept the request for Unmarshaling into T
	//  2. Unmarshal the JSON into a type wrapper, which
	//     includes type information as well as the JSON
	//     that we should use to decode into T.
	//  3. Create a new T based on the discriminator value
	//     found in (2)
	//  4. Unmarshal the "remainder" from (2) into T.
	//
	// However, when unmarshaling a Go type,
	// github.com/go-json-experiment/json looks for marshalers
	// that are registered either for that concrete type, or
	// for any interfaces which the Go type implements.
	//
	// This produces, by default, an infinite loop between
	// steps (1) and (4), where calling json.Marshal(t) will
	// re-invoke our unmarshalFunc
	//
	// To avoid this recursion, we set the skipNext toggle
	// below whenever we want a "default" unmarshaling of T.
	// If our UnmarshalFunc sees skipNext = true, it un-toggles
	// skipNext and returns [json.SkipFunc], which instructs
	// [json.Unmarshal] to skip our UnmarshalFunc.
	skipNext := false

	var unmarshalOpts *json.Unmarshalers

	unmarshalFunc := func(bytes []byte, ptr *T) error {
		// If skipNextPtr is on, toggle it off and skip this
		// custom unmarshal function. `t` will be decoded according
		// to subsequent decoding rules; including the default
		// encoding if no other rules preempt it.
		// See [json.Unmarshal].
		if skipNext {
			skipNext = false
			return json.SkipFunc
		}

		// We expect the JSON for this type to be wrapped in a
		// way that tells us what type of T we should decode into.
		//
		// So first, decode the input into our discriminated type.
		discriminated := struct {
			Type string `json:"_type"`
		}{}
		if err := json.Unmarshal(bytes, &discriminated, json.WithUnmarshalers(unmarshalOpts)); err != nil {
			return fmt.Errorf("failed to decode to type wrapper: %w", err)
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

		skipNext = true // avoid recursion in Unmarshal below
		if err := json.Unmarshal(bytes, optPtr); err != nil {
			return fmt.Errorf("failed to marshal value to option type %T: %w", opt, err)
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
			return
		}
	}
	return
}
