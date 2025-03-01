package main

import (
	"aibattle/game/world"
	"fmt"
)

func main() {
	res, err := world.RunGame(func(gs world.GameState, i int, ai world.ActionIndex) (world.UnitAction, error) {
		return world.UnitAction{}, nil
	})
	fmt.Println(err)
	fmt.Println(res)
}
