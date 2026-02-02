package cpp

import (
	"fmt"
	"log"
	"os"
)

// TODO: manage memory of pb.Type and such

// TODO: skip evaluation if batch is all-nil
func Evaluate(module Module, batch []Expression) []byte {
	// NOTE: main function is assumed not to exist

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
		statements = append(statements, Statement(Raw(s)))
	}
	addStmt(fmt.Sprintf(`std::ofstream outputFile("%s", std::ios::binary);`, outputFileName))
	addStmt(`outputFile << "[\n";`)

	for i, e := range batch {
		// TODO: just handle non-Expressions in the ast module instead
		if e == nil {
			if i < len(batch)-1 {
				addStmt(`outputFile << "\"<no_value>\"" << ",\n";`)
			} else {
				addStmt(`outputFile << "\"<no_value>\"" << "\n";`)
			}
		} else {
			if i < len(batch)-1 {
				addStmt(fmt.Sprintf(`outputFile << ty::serialize(%s) << ",\n";`, e))
			} else {
				addStmt(fmt.Sprintf(`outputFile << ty::serialize(%s) << "\n";`, e))
			}
		}
	}
	addStmt(`outputFile << "]\n";`)
	addStmt(`outputFile.close();`)

	module.Declarations = append(module.Declarations, FunctionDeclaration{
		Name:       "main",
		Parameters: []FunctionParameter{},
		ReturnType: "int",
		Body:       Block(statements),
	})
	Run(module)
	bytes, err := os.ReadFile(outputFileName)
	if err != nil {
		log.Fatalln("Failed to read output file during compile-time C++ evaluation. Error:", err)
	}
	println(string(bytes))
	return bytes // not deserialized here to prevent module import loop
}
