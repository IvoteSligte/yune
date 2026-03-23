package main

import "testing"

func BenchmarkStandardLibrary(b *testing.B) {
	for b.Loop() {
		runModuleFromFile("std.un")
	}
}
