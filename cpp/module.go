package cpp

import (
	"yune/util"
)

type Module struct {
	Declarations []TopLevelDeclaration
}

// FIXME: declarations need to be ordered properly in the header file.
// It is primarily that type declarations need to come before constant declarations that use them.

func (m Module) GenHeader() string {
	// <tuple> for std::tuple, std::apply
	// <functional> for std::function
	// <string> for std::string
	// <fstream> for std::fstream (only for evaluation right now)
	return `
#include <tuple>      // std::tuple, std::apply
#include <functional> // std::function
#include <string>     // std::string
#include <vector>     // std::vector
#include <fstream>    // std::fstream
#include <iostream>   // std::cout
` + util.JoinFunction(m.Declarations, "\n", TopLevelDeclaration.GenHeader)
}

func (m Module) String() string {
	return util.Join(m.Declarations, "\n")
}
