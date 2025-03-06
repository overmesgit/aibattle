package builder

import (
	"aibattle/game/world"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dop251/goja"
)

type GOJARunner struct {
	getTurnActions goja.Callable
	vm             *goja.Runtime
}

func NewGOJARunner(generatedCode string) (GOJARunner, error) {
	// Create a new JavaScript runtime
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// Create console object and set log method
	err := vm.Set("log", consoleLogFunc)
	if err != nil {
		return GOJARunner{}, err
	}

	_, err = vm.RunString(generatedCode)
	if err != nil {
		return GOJARunner{}, fmt.Errorf("failed to run generated code: %w", err)
	}

	getTurnActionsValue := vm.Get("GetTurnActions")
	if getTurnActionsValue == nil || goja.IsUndefined(getTurnActionsValue) {
		return GOJARunner{}, errors.New("GetTurnActions function not found in the generated code")
	}

	getTurnActions, ok := goja.AssertFunction(getTurnActionsValue)
	if !ok {
		log.Printf("GetTurnActions is not a function")
		return GOJARunner{}, errors.New("GetTurnActions is not a function")
	}
	res := GOJARunner{
		getTurnActions: getTurnActions,
		vm:             vm,
	}

	return res, nil
}

func (runner GOJARunner) GetNextAction(
	state world.GameState, unitID int, actionIndex string,
) (world.UnitAction, error) {
	res, err := runner.getTurnActions(
		goja.Undefined(), runner.vm.ToValue(state),
		runner.vm.ToValue(unitID), runner.vm.ToValue(actionIndex),
	)
	if err != nil {
		return world.UnitAction{}, fmt.Errorf("error calling GetTurnActions: %w", err)
	}
	action, err := ParseAction(res)
	if err != nil {
		return action, fmt.Errorf("error parsing action: %w", err)
	}
	return action, nil
}

func consoleLogFunc(call goja.FunctionCall) goja.Value {
	// Convert all arguments to strings and join them with a space
	var args []string
	for _, arg := range call.Arguments {
		args = append(args, fmt.Sprintf("%v", arg))
	}
	message := strings.Join(args, " ")
	log.Println("[console.log]:", message)
	return goja.Undefined()
}

func ParseAction(res goja.Value) (world.UnitAction, error) {
	// Try to parse the result into a UnitAction structure using a map approach
	action := world.UnitAction{}

	resExport := res.Export()
	// Use resExport directly to get the result as a map
	resultMap, ok := resExport.(map[string]any)
	if !ok {
		return action, errors.New(
			fmt.Sprintf(
				"result is not a map: %v (%T)", resExport, resExport,
			),
		)
	}

	// Extract action from map
	if actionVal, ok := resultMap["action"]; ok {
		actionStr, ok := actionVal.(string)
		if !ok {
			return action, errors.New(fmt.Sprintf("action is not a string: %v", actionVal))
		}
		action.Action = world.Action(actionStr)
	}

	// Extract target from map if it exists
	if targetVal, ok := resultMap["target"]; ok {
		switch targetVal.(type) {
		case *world.Position:
			action.Target = targetVal.(*world.Position)
		case map[string]any:
			position, err := parseTarget(targetVal)
			if err != nil {
				return action, err
			}
			action.Target = position
		}
	}

	return action, nil
}

func parseTarget(targetVal any) (*world.Position, error) {
	targetMap, ok := targetVal.(map[string]any)
	if !ok {
		return nil, errors.New(
			fmt.Sprintf(
				"target is not a map: %v (type: %T)", targetVal, targetVal,
			),
		)
	}
	x, xOk := targetMap["x"].(int64)
	y, yOk := targetMap["y"].(int64)
	if !xOk || !yOk {
		return nil, errors.New(
			fmt.Sprintf(
				"Warning: x or y is not a int64: %v", targetMap,
			),
		)
	}
	return &world.Position{
		X: int(x),
		Y: int(y),
	}, nil
}
