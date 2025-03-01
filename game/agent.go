package game

import (
	"aibattle/game/world"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetTurnAction(team world.Team, state []byte) ([]byte, error) {
	path := "team_one"
	if team == world.TeamB {
		path = "team_two"
	}
	url := fmt.Sprintf("http://localhost:8080/%s/", path)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(state))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code from the agent %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type NextTurnInput struct {
	State         world.GameState `json:"state"`
	CurrentUnitID int             `json:"current_unit_id"`
}

func GetNextAction(state world.GameState, unit *world.Unit) ([]world.UnitAction, error) {
	jsonState, err := json.Marshal(
		NextTurnInput{
			State:         state,
			CurrentUnitID: unit.ID,
		},
	)
	if err != nil {
		log.Println("error marshal next turn input", err)
		return nil, err
	}

	nextMoveJson, turnErr := GetTurnAction(unit.Team, jsonState)
	if turnErr != nil {
		log.Printf("team %d error getting next action %s \n", unit.Team, turnErr.Error())
		return nil, turnErr
	}

	var nextMove []world.UnitAction
	err = json.Unmarshal(nextMoveJson, &nextMove)
	if err != nil {
		log.Println("error unmarshal next move", err)
		return nextMove, err
	}
	if len(nextMove) == 0 {
		return nextMove, errors.New("no action found")
	}
	return nextMove, nil
}
