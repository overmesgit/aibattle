package builder

import (
	"aibattle/game/rules"
	"aibattle/game/world"
	"context"
	"fmt"
	"log"
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
	generatedCode, err := rules.AddGeneratedCodeToTheGameTemplate(text, language)
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
		log.Printf("Error calling getNextAction: %v", err)
		return fmt.Errorf("error calling getNextAction: %w", err)
	}
	log.Printf(
		"Successfully tested the generated code. Parsed action: %+v %+v", action, action.Target,
	)
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
