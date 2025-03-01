package world

import (
	"log"

	"github.com/samber/lo"
)

type ActionLog struct {
	Turn        int        `json:"turn,omitempty"`
	UnitID      int        `json:"unit_id"`
	UnitActions UnitAction `json:"unit_action"`
	Errors      []string   `json:"errors"`
	// TODO: fill units after
	UnitsAfter []Unit `json:"units,omitempty"`
}

type Result struct {
	Turns       []ActionLog `json:"turns"`
	Winner      int         `json:"winner"`
	InitUnits   []*Unit     `json:"init_units"`
	TeamOneLogs string      `json:"team_one_logs"`
	TeamTwoLogs string      `json:"team_two_logs"`
}

func (r Result) NewActionLog(turn int, unitID int) ActionLog {
	turnLog := ActionLog{
		Turn:   turn,
		UnitID: unitID,
	}
	r.Turns = append(r.Turns, turnLog)
	return turnLog
}

type UnitAction struct {
	Action Action    `json:"action"`
	Target *Position `json:"target"`
}

type ActionIndex string

var FirstAction ActionIndex = "FirstAction"
var SecondAction ActionIndex = "SecondAction"

func RunGame(
	nextAction func(
		int, GameState, int, ActionIndex,
	) (UnitAction, error),
) (Result, error) {
	gameState := GetInitialGameState()
	maxTurns := 50

	teamA := lo.Filter(
		gameState.Units, func(item *Unit, _ int) bool { return item.Team == TeamA },
	)
	teamB := lo.Filter(
		gameState.Units, func(item *Unit, _ int) bool { return item.Team == TeamB },
	)
	result := Result{
		Winner:    Draw,
		InitUnits: gameState.Units,
	}

	for turn := range maxTurns {
		gameState.Turn = turn

		log.Printf(
			"Turn %d TeamA %d TeamB %d\n", turn, lo.SumBy(teamA, calcHP), lo.SumBy(teamB, calcHP),
		)

		wonTeam, gameOver := checkWinningTeam(teamA, gameState, teamB)
		if gameOver {
			result.Winner = wonTeam
			break
		}

		for _, unit := range gameState.Units {
			if !unit.IsAlive() {
				continue
			}

			prevAction := Action("")
			for _, actIndex := range []ActionIndex{FirstAction, SecondAction} {
				actionLog := result.NewActionLog(turn, unit.ID)

				act, actionErr := nextAction(unit.Team, gameState, unit.ID, actIndex)
				log.Printf("next action %v %+v %+v %v", unit.ID, act, act.Target, actionErr)
				if actionErr != nil {
					actionLog.Errors = append(actionLog.Errors, actionErr.Error())
					continue
				}
				err := gameState.UpdateGameState(*unit, act, prevAction)
				if err != nil {
					actionLog.Errors = append(actionLog.Errors, err.Error())
					log.Println(err)
				}
				prevAction = act.Action
			}
		}
		gameState.RemoveDeadUnits()
	}
	return result, nil
}

func checkWinningTeam(
	teamA []*Unit, gameState GameState, teamB []*Unit,
) (int, bool) {
	teamA = lo.Filter(gameState.Units, AliveTeamUnits(TeamA))
	if len(teamA) == 0 {
		log.Println("Team B wins!")
		return TeamB, true
	}
	teamB = lo.Filter(gameState.Units, AliveTeamUnits(TeamB))
	if len(teamB) == 0 {
		log.Println("Team A wins!")
		return TeamA, true
	}
	return -1, false
}

func AliveTeamUnits(team int) func(unit *Unit, index int) bool {
	return func(unit *Unit, index int) bool {
		return unit.IsAlive() && unit.Team == team
	}
}

func calcHP(item *Unit) int {
	return item.HP
}
