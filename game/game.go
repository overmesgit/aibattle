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
	State  world.GameState `json:"state"`
	UnitID int             `json:"unit_id"`
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetNextAction(state world.GameState, unit *world.Unit) TurnAction {
	jsonState, err := json.Marshal(
		NextTurnInput{
			State:  state,
			UnitID: unit.ID,
		},
	)
	if err != nil {
		panic(err)
	}

	nextMoveJson, err := GetTurnAction(unit.Team, jsonState)

	nextMove := TurnAction{
		UnitID: unit.ID,
	}
	err = json.Unmarshal(nextMoveJson, &nextMove)
	if err != nil {
		log.Println(err)
	}
	return nextMove
}

type Result struct {
	Turns  []TurnLog
	Winner int
}

func RunGame() (Result, error) {
	gameState := world.GetInitialGameState()
	maxTurns := 50

	var teamA []*world.Unit
	var teamB []*world.Unit
	calcHP := func(item *world.Unit) int {
		return item.HP
	}
	result := Result{
		Winner: world.Draw,
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
				log.Printf("Team %d Unit %s is dead\n", unit.Team, unit.Type)
				continue
			}

			nextAction := GetNextAction(gameState, unit)
			turnActions = append(turnActions, nextAction)

			err := checkActionsAreUnique(nextAction)
			if err != nil {
				log.Println(err)
				nextAction.UnitAction[1].Error = err.Error()
				continue
			}
			for _, act := range nextAction.UnitAction {
				if act == nil || act.Action == "" {
					continue
				}
				//log.Printf(
				//	"Team %d Unit %s %v performs %v %v\n", unit.Team, unit.Type, unit.Position,
				//	act.Action,
				//	act.Target,
				//)
				err := updateGameState(&gameState, unit, act)
				if err != nil {
					log.Println(err)
					act.Error = err.Error()
				}
			}

		}
		turnLog.Actions = turnActions
		result.Turns = append(result.Turns, turnLog)
	}
	return result, nil
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

func checkActionsAreUnique(nextAction TurnAction) error {
	actions := lo.Map(
		nextAction.UnitAction[:], func(item *UnitAction, index int) world.Action {
			if item == nil {
				return ""
			}
			return item.Action
		},
	)
	if len(actions) > 1 && actions[0] == actions[1] {
		return errors.New("unit can perform only one same action per turn")
	}
	return nil
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
