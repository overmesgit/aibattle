package builder

import (
	"aibattle/game/rules"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func GetProgram(
	ctx context.Context, promptID string, prompt string, language string,
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
	buildErr := buildImage(generatedCode, promptID, language)
	if buildErr != nil {
		return text, fmt.Errorf("error building image %s", buildErr)
	}
	return text, nil
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

func buildImage(generatedCode string, promptID string, language string) error {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "aibattle-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}

	var progErr error
	if language == rules.LangGo {
		progErr = GetGoProgram(generatedCode, tmpDir)
	} else if language == rules.LangPy {
		progErr = GetPyProgram(generatedCode, tmpDir)
	} else {
		progErr = errors.New("unknown language")
	}
	if progErr != nil {
		return progErr
	}

	log.Println("Building image")
	errBuild := BuildImageInTmpFolder(tmpDir, promptID)
	if errBuild != nil {
		return errBuild
	}
	return nil
}

func AddGeneratedCodeToTheGameTemplate(txt string, language string) (string, error) {
	mainTemplate := ""
	if language == rules.LangGo {
		mainTemplate = "game/rules/templates/go_test.go"
	} else if language == rules.LangPy {
		mainTemplate = "game/rules/templates/py.py"
	}
	tfile, err := os.ReadFile(mainTemplate)
	if err != nil {
		return "", err
	}

	// Replace <generated> with provided text
	template := strings.Replace(string(tfile), "<generated>", txt, 1)
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

func BuildImageInTmpFolder(tmpDir string, promptID string) error {
	// Build the Docker image
	log.Printf("Building image in tmp folder %s", tmpDir)
	imageTag := fmt.Sprintf("ai%s:latest", promptID)
	cmd := exec.Command("docker", "build", "-t", imageTag, tmpDir)
	var stdout strings.Builder
	cmd.Stdout = &stdout

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(
			"failed to build docker image: %v\nstderr: %s\noutput: %s", err, stderr.String(),
			stdout.String(),
		)
	}

	return nil
}
