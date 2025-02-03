package main

import (
	"aibattle/pages/prompt"
	"fmt"
)

func main() {
	result, err := prompt.AddGeneratedCodeToTheGameTemplate("test")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result)
}
