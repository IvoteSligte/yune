package pb

import (
	"fmt"
	"log"
	"os"
	"yune/cpp"

	"capnproto.org/go/capnp/v3"
)

// TODO: skip evaluation if batch is all-nil
func Evaluate(module cpp.Module, batch []cpp.Expression) (outputs []Value) {
	// NOTE: main function is assumed not to exist

	fmt.Println("--- Start Evaluation ---")
	defer fmt.Println("--- End Evaluation ---")

	outputFile, err := os.CreateTemp("", "yune-eval")
	if err != nil {
		log.Fatalln("Failed to create temporary file during compile-time C++ evaluation. Error:", err)
	}
	defer os.Remove(outputFile.Name()) // TODO: close file

	statements := []cpp.Statement{}

	addStmt := func(s string) {
		statements = append(statements, cpp.Statement(cpp.Raw(s)))
	}
	addStmt(fmt.Sprintf(`std::ofstream outputFile("%s", std::ios::binary);`, outputFile.Name()))

	for _, e := range batch {
		addStmt(`::capnp::MallocMessage:Builder message;`)
		addStmt(`pb::Value value = message.initRoot<pb::Value>();`)
		if e != nil {
			statement = `outputFile << ` + e.String() + ` << '\0';`
		}
		addStmt(`::capnp::writeMessage(outputFile, message);`)
	}
	statements = append(statements, cpp.Statement(cpp.Raw(`outputFile.close();`)))

	module.Declarations = append(module.Declarations, cpp.FunctionDeclaration{
		Name:       "main",
		Parameters: []cpp.FunctionParameter{},
		ReturnType: "int",
		Body:       cpp.Block(statements),
	})
	cpp.Run(module)
	decoder := capnp.NewDecoder(outputFile)
	for range len(batch) {
		msg, err := decoder.Decode()
		if err != nil {
			log.Fatalf("Failed to decode compile-time evaluation outputs. Error: %s", err)
		}
		value, err := ReadRootValue(msg)
		if err != nil {
			log.Fatalf("Failed to parse compile-time evaluation outputs. Error: %s", err)
		}
		outputs = append(outputs, value)
	}
	return
}
