package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

type Result struct {
	*Ast   `json:"ast"`
	Source string `json:"source"`
	Dump   string `json:"dump"`
}

func main() {
	src := js.Global().Get("source")
	source := src.String()
	ast, dump, err := Parse("foo", source)
	if err != nil {
		fmt.Println("Error", err)
	}
	result := Result{Ast: ast, Source: source, Dump: dump}
	body, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error", err)
	}
	js.Global().Set("output", string(body))
}
