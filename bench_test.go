package main

import "testing"

// Parse-only benchmark for the standard library
func BenchmarkParseStandardLibrary(b *testing.B) {
	fileName := "std.un"
	sourceCode := readFile(fileName)
	for b.Loop() {
		parseModule(fileName, sourceCode)
	}
}

// Parse+analyse benchmark for the standard library
func BenchmarkCompileStandardLibrary(b *testing.B) {
	for b.Loop() {
		runModuleFromFile("std.un")
	}
}
