module yune/main

go 1.25.2

require github.com/antlr4-go/antlr/v4 v4.13.1

require (
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
)

// TODO: github.com permalink instead of local
replace github.com/go-json-experiment/json => ../json