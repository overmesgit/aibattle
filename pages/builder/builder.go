package builder

import (
	"aibattle/game/rules"
	"aibattle/game/world"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
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
	err = RunGOJACodeTest(generatedCode)
	if err != nil {
		return text, err
	}
	return text, nil
}

func RunGOJACodeTest(generatedCode string) error {
	gameState := world.GetInitialGameState()

	getNextAction, err := GetGOJAFunction(generatedCode)
	if err != nil {
		log.Printf("Error preparing js function: %v", err)
		return fmt.Errorf("error preparing js function: %w", err)
	}

	action, err := getNextAction(
		gameState, 1, "FirstAction",
	)
	if err != nil {
		log.Printf("Error calling GetTurnActions: %v", err)
		return fmt.Errorf("error calling GetTurnActions: %w", err)
	}
	log.Printf("Successfully tested the generated code. Parsed action: %+v", action)
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
