package builder

import (
	"aibattle/game/world"
	"encoding/json"
	"fmt"

	"github.com/buke/quickjs-go"
)

type QuickJSRunner struct {
	ctx *quickjs.Context
}

func NewQuickJSRunner(generatedCode string) (QuickJSRunner, error) {
	// Create a new QuickJS runtime and context
	runtime := quickjs.NewRuntime(
		quickjs.WithExecuteTimeout(3),
		quickjs.WithMemoryLimit(5*1024*1024),
		quickjs.WithGCThreshold(256*1024),
		quickjs.WithMaxStackSize(65534),
	)
	ctx := runtime.NewContext()

	// Execute the generated code
	result, err := ctx.Eval(generatedCode)
	if err != nil {
		return QuickJSRunner{}, fmt.Errorf("failed to run generated code: %w", err)
	}
	defer result.Free()

	return QuickJSRunner{
		ctx: ctx,
	}, nil
}

func (runner QuickJSRunner) GetNextAction(
	state world.GameState, unitID int, actionIndex string,
) (world.UnitAction, error) {
	// Convert Go values to JSON strings
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return world.UnitAction{}, fmt.Errorf("error marshaling state to JSON: %w", err)
	}

	// Create JS values from JSON
	stateJSValue := runner.ctx.ParseJSON(string(stateJSON))
	if stateJSValue.IsException() {
		exception := runner.ctx.Exception()
		return world.UnitAction{}, fmt.Errorf("error parsing state JSON: %w", exception)
	}
	defer stateJSValue.Free()

	unitIDJSValue := runner.ctx.Int32(int32(unitID))
	defer unitIDJSValue.Free()

	actionIndexJSValue := runner.ctx.String(actionIndex)
	defer actionIndexJSValue.Free()

	// Call the JS function
	result := runner.ctx.Globals().Call(
		"GetTurnActions", stateJSValue, unitIDJSValue, actionIndexJSValue,
	)
	if result.IsException() {
		exception := runner.ctx.Exception()
		return world.UnitAction{}, fmt.Errorf(
			"exception when calling GetTurnActions: %w", exception,
		)
	}
	defer result.Free()

	// Convert result back to Go
	resultJSON := result.JSONStringify()

	var action world.UnitAction
	err = json.Unmarshal([]byte(resultJSON), &action)
	if err != nil {
		return world.UnitAction{}, fmt.Errorf("error unmarshaling action from JSON: %w", err)
	}

	return action, nil
}
