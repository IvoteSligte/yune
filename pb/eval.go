package pb

import (
	"fmt"
	"log"
	"os"
	"yune/cpp"
)

// TODO: manage memory of pb.Type and such

// TODO: skip evaluation if batch is all-nil
func Evaluate(module cpp.Module, batch []cpp.Expression) (outputs []Value) {
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

	statements := []cpp.Statement{}

	addStmt := func(s string) {
		statements = append(statements, cpp.Statement(cpp.Raw(s)))
	}
	addStmt(fmt.Sprintf(`std::ofstream outputFile("%s", std::ios::binary);`, outputFileName))
	addStmt(`std::vector<Value> outputs;`)

	for _, e := range batch {
		if e == nil {
			addStmt(`outputs.emplace_back({});`)
		} else {
			addStmt(fmt.Sprintf(`outputs.push_back(%s);`, e))
		}
	}
	addStmt(`outputFile << serializeValues(outputs);`)
	addStmt(`outputFile.close();`)

	module.Declarations = append(module.Declarations, cpp.FunctionDeclaration{
		Name:       "main",
		Parameters: []cpp.FunctionParameter{},
		ReturnType: "int",
		Body:       cpp.Block(statements),
	})
	cpp.Run(module)
	bytes, err := os.ReadFile(outputFileName)
	if err != nil {
		log.Fatalln("Failed to read output file during compile-time C++ evaluation. Error:", err)
	}
	values := DeserializeValues(string(bytes))
	return ToSlice(values)
}
