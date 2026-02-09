package cpp

import (
	"yune/util"
)

type Module struct {
	Declarations []Declaration
}

func (m Module) Header() string {
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
` + util.JoinFunction(m.Declarations, "\n\n", func(d Declaration) string {
		return d.Header
	})
}

func (m Module) Implementation() string {
	return util.JoinFunction(m.Declarations, "\n\n", func(d Declaration) string {
		return d.Implementation
	})
}
