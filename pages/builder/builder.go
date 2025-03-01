package builder

import (
	"aibattle/game/rules"
	"aibattle/game/world"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/dop251/goja"
)

func GetProgram(
	ctx context.Context, prompt string, language string,
) (string, error) {
	gameRules, err := rules.GetGameDescription(language)
	if err != nil {
		return "", err
	}
	text, promptErr := GetProgramWithPrompt(ctx, prompt, gameRules)
	if promptErr != nil {
		return text, promptErr
	}

	// Get the generated code
	generatedCode, err := AddGeneratedCodeToTheGameTemplate(text, language)
	if err != nil {
		return text, err
	}
	err = RunGojaCodeTest(generatedCode)
	if err != nil {
		return text, err
	}
	return text, nil
}

func RunGojaCodeTest(generatedCode string) error {
	// Create a new JavaScript runtime
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// Define a console.log function and pass it to the JS runtime
	consoleLog := func(call goja.FunctionCall) goja.Value {
		// Convert all arguments to strings and join them with a space
		var args []string
		for _, arg := range call.Arguments {
			args = append(args, fmt.Sprintf("%v", arg))
		}
		message := strings.Join(args, " ")
		log.Println("[console.log]:", message)
		return goja.Undefined()
	}

	// Create console object and set log method
	err := vm.Set("log", consoleLog)
	if err != nil {
		return err
	}

	_, err = vm.RunString(generatedCode)
	if err != nil {
		log.Printf("Error running generated code: %v", err)
		return fmt.Errorf("failed to run generated code: %w", err)
	}

	getTurnActionsValue := vm.Get("GetTurnActions")
	if getTurnActionsValue == nil || goja.IsUndefined(getTurnActionsValue) {
		log.Printf("GetTurnActions function not found in the generated code")
		return errors.New("GetTurnActions function not found in the generated code")
	}

	getTurnActions, ok := goja.AssertFunction(getTurnActionsValue)
	if !ok {
		log.Printf("GetTurnActions is not a function")
		return errors.New("GetTurnActions is not a function")
	}

	// Create a mock game state for testing
	gameState := world.GetInitialGameState()
	// Call the function with mock values
	res, err := getTurnActions(
		goja.Undefined(), vm.ToValue(gameState),
		vm.ToValue(1), vm.ToValue("FirstAction"),
	)
	if err != nil {
		log.Printf("Error calling GetTurnActions: %v", err)
		return fmt.Errorf("error calling GetTurnActions: %w", err)
	}

	log.Printf("Test ouput %v\n", res.Export())

	// Try to parse the result into a UnitAction structure using a map approach
	action := world.UnitAction{}

	// Use res.Export() directly to get the result as a map
	resultMap, ok := res.Export().(map[string]any)
	if !ok {
		log.Printf("Warning: Result is not a map: %v", res.Export())
		log.Printf("Raw result: %#v", res.Export())
	} else {
		// Extract action from map
		if actionVal, ok := resultMap["action"]; ok {
			if actionStr, ok := actionVal.(string); ok {
				action.Action = world.Action(actionStr)
			}
		}
		// Extract target from map if it exists
		if targetVal, ok := resultMap["target"]; ok {
			if targetMap, ok := targetVal.(map[string]any); ok {
				x, xOk := targetMap["x"].(int64)
				y, yOk := targetMap["y"].(int64)
				if xOk && yOk {
					action.Target = &world.Position{
						X: int(x),
						Y: int(y),
					}
				}
			}
		}
		log.Printf("Successfully tested the generated code. Parsed action: %+v", action)
	}
	return nil
}

func GetProgramWithPrompt(
	ctx context.Context, prompt string, rules string,
) (string, error) {
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	}
	// defaults to os.LookupEnv("ANTHROPIC_API_KEY")
	client := anthropic.NewClient()
	resp, err := client.Messages.New(
		ctx, anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
			MaxTokens: anthropic.Int(8192),
			System: anthropic.F(
				[]anthropic.TextBlockParam{
					anthropic.NewTextBlock(rules),
				},
			),
			Messages: anthropic.F(messages),
		},
	)
	if err != nil || resp == nil {
		log.Println(err)
		return "", err
	}
	text := resp.Content[0].Text
	log.Printf("%+v\n", text[:100])
	startTag := "<sourcecode>"
	endTag := "</sourcecode>"
	text, err = getContentBetweenTags(text, startTag, endTag)
	if err != nil {
		return "", err
	}
	return text, nil
}

func AddGeneratedCodeToTheGameTemplate(txt string, language string) (string, error) {
	mainTemplate := ""
	switch language {
	case rules.LangGo:
		mainTemplate = "game/rules/templates/go_test.go"
	case rules.LangPy:
		mainTemplate = "game/rules/templates/py.py"
	case rules.LangJS:
		mainTemplate = "game/rules/templates/js.js"
	}
	tfile, err := os.ReadFile(mainTemplate)
	if err != nil {
		return "", err
	}

	strContent := string(tfile)
	getTagCount := strings.Count(strContent, "<generated>")
	if getTagCount == 0 || getTagCount > 1 {
		return "", fmt.Errorf("no or too many <generated> tags found")
	}

	// Replace <generated> with provided text
	template := strings.Replace(strContent, "<generated>", txt, 1)
	return template, nil
}

func getContentBetweenTags(content, startTag, endTag string) (string, error) {
	startIdx := strings.LastIndex(content, startTag) + len(startTag)
	endIdx := strings.LastIndex(content, endTag)
	if startIdx == -1 || endIdx == -1 {
		return "", fmt.Errorf("tags not found: %s, %s", startTag, endTag)
	}
	tagText := content[startIdx:endIdx]
	if len(tagText) == 0 {
		return "", fmt.Errorf("text between tags are empty")
	}
	return tagText, nil
}
