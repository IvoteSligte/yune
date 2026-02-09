package cpp

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// TODO: manage memory of pb.Type and such

// TODO: skip evaluation if batch is all-nil
func Evaluate(module Module, batch []Expression) []string {
	// NOTE: main function is assumed not to exist and is ignored if it does

	fmt.Println("--- Start Evaluation ---")
	defer fmt.Println("--- End Evaluation ---")

	outputFile, err := os.CreateTemp("", "yune-eval")
	if err != nil {
		log.Fatalln("Failed to create temporary file during compile-time C++ evaluation. Error:", err)
	}
	outputFileName := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		log.Fatalln("Failed to close temporary file during compile-time C++ evaluation. Error:", err)
	}
	defer os.Remove(outputFileName)

	statements := []Statement{}

	addStmt := func(s string) {
		statements = append(statements, s)
	}
	addStmt(fmt.Sprintf(`std::ofstream outputFile("%s", std::ios::binary);`, outputFileName))

	for _, e := range batch {
		// data separated by newlines (newlines are escaped in string literals)
		if e == "" {
			addStmt(`outputFile << "\n";`) // no data
		} else {
			addStmt(fmt.Sprintf(`outputFile << ty::serialize(%s) << "\n";`, e))
		}
	}
	addStmt(`outputFile.close();`)
	addStmt(`return 0;`)

	module.Declarations = append(module.Declarations, Declaration{
		Header: "",
		Implementation: fmt.Sprintf(`int main() {
    %s
    return 0;
}`, strings.Join(statements, "\n")),
	})
	Run(module)
	bytes, err := os.ReadFile(outputFileName)
	if err != nil {
		log.Fatalln("Failed to read output file during compile-time C++ evaluation. Error:", err)
	}
	println(string(bytes))
	evalJsons := strings.Split(string(bytes), "\n") // not deserialized here to prevent module import loop
	return evalJsons[:len(evalJsons)-1]             // skip trailing line
}
