package ast

import "yune/cpp"

// check for duplicate declarations
// infer and check types (also of statement body)

// Lowers the root module.
func LowerModule(m Module) cpp.Module {
	env := Environment{
		parent: nil,
		// TODO: builtins
		symbols: map[string]Declaration{},
	}
	return env.LowerModule(m)
}

func (env *Environment) LowerModule(m Module) cpp.Module {
	panic("unimplemented")
}
