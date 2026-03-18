package ast

import (
	"fmt"
	"regexp"
)

// Stores global state that needs to be retained between Analyze and Lower steps.
type State struct {
	// Stores closures that need to be serializable from C++.
	registeredClosures map[string]*Closure
	// Stores type values that need to be serializable from C++.
	registeredTypeValues map[string]TypeValue
}

func NewState() *State {
	return &State{
		registeredClosures:   map[string]*Closure{},
		registeredTypeValues: map[string]TypeValue{},
	}
}

var invalidIdentifierChar = regexp.MustCompile("[^a-zA-Z_0-9]")

// Maps an arbitrary string to a valid C identifier, except that it may also be a C keyword.
func stringToIdentifier(name string) (result string) {
	for _, c := range name {
		if invalidIdentifierChar.MatchString(string(c)) {
			result += "_"
		} else {
			result += string(c)
		}
	}
	return
}

func (s *State) registerClosure(closure *Closure) string {
	// NOTE: this requires unique Span for C++-generated Closure definitions
	id := fmt.Sprintf("closure_%s_%d_%d", stringToIdentifier(closure.Span.File), closure.Span.Line, closure.Span.Column)
	s.registeredClosures[id] = closure
	return id
}

func (s *State) registerTypeValue(typeValue TypeValue) string {
	id := typeValue.String()
	s.registeredTypeValues[id] = typeValue
	return id
}
