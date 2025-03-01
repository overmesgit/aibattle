package main

import (
	"aibattle/game/rules"
	"fmt"
)

func main() {
	rules, err := rules.GetGameDescription(rules.LangJS)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rules)

}
