package main

import (
	"aibattle/game"
	"fmt"
)

func main() {
	res, err := game.RunGame()
	fmt.Println(err)
	fmt.Println(res)
}
