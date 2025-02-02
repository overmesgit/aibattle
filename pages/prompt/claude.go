package prompt

import (
	"aibattle/game/rules"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func GetProgram(ctx context.Context, promptID string, prompt string) (string, error) {
	gameRules, err := rules.GetGameDescription()
	if err != nil {
		return "", err
	}
	text, promptErr := GetProgramWithPrompt(ctx, prompt, err, gameRules)
	if promptErr != nil {
		return text, promptErr
	}

	// Get the generated code
	generatedCode, err := ReplaceGeneratedInRulesTemplate(text)
	if err != nil {
		return text, err
	}
	buildErr := buildImage(err, generatedCode, promptID)
	if buildErr != nil {
		return text, fmt.Errorf("error building image %s", buildErr)
	}
	return text, nil
}

func GetProgramWithPrompt(
	ctx context.Context, prompt string, err error, rules string,
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

func buildImage(err error, generatedCode string, promptID string) error {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "aibattle-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}

	tmpDir, errMain := CreateMainGoInTmpFolder(tmpDir, generatedCode)
	if errMain != nil {
		return errMain
	}
	errBuild := BuildImageInTmpFolder(tmpDir, promptID)
	if errBuild != nil {
		return errBuild
	}
	return nil
}

func ReplaceGeneratedInRulesTemplate(txt string) (string, error) {
	tfile, err := os.ReadFile("game/rules/rules.tmpl")
	if err != nil {
		return "", err
	}

	content := string(tfile)

	startTag := "<template>"
	endTag := "</template>"
	template, err := getContentBetweenTags(content, startTag, endTag)
	if err != nil {
		return "", err
	}
	// Replace <generated> with provided text
	template = strings.Replace(template, "<generated>", txt, 1)

	return template, nil
}

func getContentBetweenTags(content, startTag, endTag string) (string, error) {
	startIdx := strings.LastIndex(content, startTag) + len(startTag)
	endIdx := strings.LastIndex(content, endTag)
	if startIdx == -1 || endIdx == -1 {
		return "", fmt.Errorf("tags not found: %s, %s", startTag, endTag)
	}
	return content[startIdx:endIdx], nil
}

func CreateMainGoInTmpFolder(tmpDir string, generatedCode string) (string, error) {
	// Create main.go file
	mainPath := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainPath, []byte(generatedCode), 0644)
	if err != nil {
		return "", err
	}

	// Create Dockerfile
	dockerfile := `FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init aibattle && go build -o main

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]`

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return tmpDir, nil
}

func BuildImageInTmpFolder(tmpDir string, promptID string) error {
	// Build the Docker image
	imageTag := fmt.Sprintf("ai%s:latest", promptID)
	cmd := exec.Command("docker", "build", "-t", imageTag, tmpDir)
	cmd.Stdout = os.Stdout

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build docker image: %v\nstderr: %s", err, stderr.String())
	}

	return nil
}
