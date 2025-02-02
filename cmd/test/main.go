package main

import (
	"aibattle/pages/prompt"
	"fmt"
)

func main() {
	result, err := prompt.ReplaceGeneratedInRulesTemplate("test")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result)
}
