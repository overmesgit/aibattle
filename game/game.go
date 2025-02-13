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

	"github.com/samber/lo"
)

func updateGameState(gameState *world.GameState, unit *world.Unit, action *UnitAction) error {
	if !unit.IsAlive() {
		return errors.New(fmt.Sprintf("Unit %d is not alive", unit.ID))
	}
	// TODO: check move through

	switch action.Action {
	case world.HOLD:
		return nil
	case world.MOVE:
		err := gameState.MoveUnit(unit, action.Target)
		if err != nil {
			return err
		}
	case world.ATTACK1:
		err := gameState.AttackUnit(unit, action.Target)
		if err != nil {
			return err
		}
	case world.SKILL1:
		err := gameState.UseSkill(unit, action.Target)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unknown action %s", action.Action))
	}
	return nil
}

type TurnAction struct {
	UnitAction [2]*UnitAction `json:"unit_action"`
	UnitID     int            `json:"unit_id"`
}

type UnitAction struct {
	Action world.Action    `json:"action"`
	Target *world.Position `json:"target"`
	Error  string          `json:"error"`
}

type NextTurnInput struct {
	State         world.GameState `json:"state"`
	CurrentUnitID int             `json:"current_unit_id"`
}

func GetTurnAction(team int, state []byte) ([]byte, error) {
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

func GetNextAction(state world.GameState, unit *world.Unit) (TurnAction, error) {
	jsonState, err := json.Marshal(
		NextTurnInput{
			State:         state,
			CurrentUnitID: unit.ID,
		},
	)
	if err != nil {
		log.Println("error marshal next turn input", err)
		return TurnAction{}, err
	}

	nextMoveJson, turnErr := GetTurnAction(unit.Team, jsonState)
	if turnErr != nil {
		log.Println("error getting next action", turnErr)
		return TurnAction{}, turnErr
	}

	nextMove := TurnAction{
		UnitID: unit.ID,
	}
	err = json.Unmarshal(nextMoveJson, &nextMove)
	if err != nil {
		log.Println("error unmarshal next move", err)
		return nextMove, err
	}
	if nextMove.UnitAction[0] == nil && nextMove.UnitAction[1] == nil {
		return nextMove, errors.New("no action found")
	}
	return nextMove, nil
}

type Result struct {
	Turns       []TurnLog     `json:"turns"`
	Winner      int           `json:"winner"`
	InitUnits   []*world.Unit `json:"init_units"`
	TeamOneLogs string        `json:"team_one_logs"`
	TeamTwoLogs string        `json:"team_two_logs"`
}

func RunGame() (Result, error) {
	gameState := world.GetInitialGameState()
	maxTurns := 50

	calcHP := func(item *world.Unit) int {
		return item.HP
	}
	teamA := lo.Filter(
		gameState.Units, func(item *world.Unit, _ int) bool { return item.Team == world.TeamA },
	)
	teamB := lo.Filter(
		gameState.Units, func(item *world.Unit, _ int) bool { return item.Team == world.TeamB },
	)
	result := Result{
		Winner:    world.Draw,
		InitUnits: gameState.Units,
	}

	for turn := 0; turn < maxTurns; turn++ {
		gameState.Turn = turn
		fmt.Printf(
			"Turn %d TeamA %d TeamB %d\n", turn, lo.SumBy(teamA, calcHP), lo.SumBy(teamB, calcHP),
		)

		turnActions := make([]TurnAction, 0)
		turnLog := TurnLog{
			Turn:  gameState.Turn,
			Units: gameState.CopyUnits(),
			Type:  "turn",
		}

		wonTeam, gameOver := checkWinningTeam(teamA, gameState, teamB)
		if gameOver {
			turnLog.Type = world.GetTeamName(wonTeam)
			result.Turns = append(result.Turns, turnLog)
			result.Winner = wonTeam
			break
		}

		for _, unit := range gameState.Units {
			if !unit.IsAlive() {
				//log.Printf("Team %d Unit %s is dead\n", unit.Team, unit.Type)
				continue
			}

			nextAction, actionErr := GetNextAction(gameState, unit)
			if actionErr != nil {
				nextAction.UnitAction = [2]*UnitAction{
					{
						Error: fmt.Sprintf(
							"error reading response from the container: %s", actionErr.Error(),
						)},
				}
				continue
			}
			turnActions = append(turnActions, nextAction)

			lastAction := world.Action("")
			for _, act := range nextAction.UnitAction {
				if act == nil || act.Action == "" {
					continue
				}
				if lastAction != "" && actionsAreOk(lastAction, act.Action) {
					act.Error = "same type of actions as first action"
					continue
				}
				err := updateGameState(&gameState, unit, act)
				if err != nil {
					//log.Println(err)
					act.Error = err.Error()
				}
				lastAction = act.Action
			}

		}
		gameState.Units = lo.Filter(
			gameState.Units, func(unit *world.Unit, index int) bool {
				return unit.IsAlive()
			},
		)
		turnLog.Actions = turnActions
		result.Turns = append(result.Turns, turnLog)
	}
	return result, nil
}

var moveActions = []world.Action{world.MOVE, world.HOLD}

func actionsAreOk(prevActin world.Action, nextAction world.Action) bool {
	if lo.IndexOf(moveActions, prevActin) >= 0 && lo.IndexOf(moveActions, nextAction) >= 0 {
		return true
	}
	return false
}

func checkWinningTeam(
	teamA []*world.Unit, gameState world.GameState, teamB []*world.Unit,
) (int, bool) {
	teamA = lo.Filter(gameState.Units, AliveTeamUnits(world.TeamA))
	if len(teamA) == 0 {
		fmt.Println("Team B wins!")
		return world.TeamB, true
	}
	teamB = lo.Filter(gameState.Units, AliveTeamUnits(world.TeamB))
	if len(teamB) == 0 {
		fmt.Println("Team A wins!")
		return world.TeamA, true
	}
	return -1, false
}

type TurnLog struct {
	Turn    int          `json:"turn,omitempty"`
	Units   []world.Unit `json:"units,omitempty"`
	Actions []TurnAction `json:"actions"`
	Type    string       `json:"type"`
}

func AliveTeamUnits(team int) func(unit *world.Unit, index int) bool {
	return func(unit *world.Unit, index int) bool {
		return unit.IsAlive() && unit.Team == team
	}
}
