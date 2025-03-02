package builder

import (
	"aibattle/game/world"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dop251/goja"
)

func GetGOJAFunction(generatedCode string) (func(
	world.GameState, int, string,
) (world.UnitAction, error), error) {
	// Create a new JavaScript runtime
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// Create console object and set log method
	err := vm.Set("log", consoleLogFunc)
	if err != nil {
		return nil, err
	}

	_, err = vm.RunString(generatedCode)
	if err != nil {
		return nil, fmt.Errorf("failed to run generated code: %w", err)
	}

	getTurnActionsValue := vm.Get("GetTurnActions")
	if getTurnActionsValue == nil || goja.IsUndefined(getTurnActionsValue) {
		return nil, errors.New("GetTurnActions function not found in the generated code")
	}

	getTurnActions, ok := goja.AssertFunction(getTurnActionsValue)
	if !ok {
		log.Printf("GetTurnActions is not a function")
		return nil, errors.New("GetTurnActions is not a function")
	}

	return func(
		state world.GameState, unitID int, actionIndex string,
	) (world.UnitAction, error) {
		res, err := getTurnActions(
			goja.Undefined(), vm.ToValue(state),
			vm.ToValue(unitID), vm.ToValue(actionIndex),
		)
		if err != nil {
			return world.UnitAction{}, fmt.Errorf("error calling GetTurnActions: %w", err)
		}
		action, err := ParseAction(res)
		if err != nil {
			return action, fmt.Errorf("error parsing action: %w", err)
		}
		return action, nil
	}, nil
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

	// Use res.Export() directly to get the result as a map
	resultMap, ok := res.Export().(map[string]any)
	if !ok {
		return action, errors.New(fmt.Sprintf("Warning: Result is not a map: %v", res.Export()))
	} else {
		// Extract action from map
		if actionVal, ok := resultMap["action"]; ok {
			actionStr, ok := actionVal.(string)
			if !ok {
				return action, errors.New(
					fmt.Sprintf(
						"Warning: Result is not a string: %v", actionVal,
					),
				)
			}
			action.Action = world.Action(actionStr)
		}
		// Extract target from map if it exists
		if targetVal, ok := resultMap["target"]; ok {
			targetMap, ok := targetVal.(map[string]any)
			if !ok {
				return action, errors.New(
					fmt.Sprintf(
						"Warning: Result is not a map: %v", targetVal,
					),
				)
			}
			x, xOk := targetMap["x"].(int64)
			y, yOk := targetMap["y"].(int64)
			if !xOk || !yOk {
				return action, errors.New(
					fmt.Sprintf(
						"Warning: Result is not a int64: %v", targetMap,
					),
				)
			}
			action.Target = &world.Position{
				X: int(x),
				Y: int(y),
			}
		}

	}
	return action, nil
}
