package cpp

import (
	"strings"
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
	prefix := `
#include <tuple>      // std::tuple, std::apply
#include <functional> // std::function
#include <string>     // std::string
#include <vector>     // std::vector
#include <fstream>    // std::fstream
#include <iostream>   // std::cout

// TODO: declare Type via ast/builtin.go?
struct Type {
    std::string id;
};

std::ostream& operator<<(std::ostream& out, const Type& t) {
    return out << t.id;
}
`
	return prefix + strings.Join(util.Map(m.Declarations, TopLevelDeclaration.GenHeader), "\n")
}

func (m Module) String() string {
	return util.Join(m.Declarations, "\n")
}
