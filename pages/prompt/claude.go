package prompt

import (
	"aibattle/game/rules"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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
	text, promptErr := GetProgramWithPrompt(ctx, prompt, gameRules)
	if promptErr != nil {
		return text, promptErr
	}

	// Get the generated code
	generatedCode, err := AddGeneratedCodeToTheGameTemplate(text)
	if err != nil {
		return text, err
	}
	buildErr := buildImage(generatedCode, promptID)
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

func buildImage(generatedCode string, promptID string) error {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "aibattle-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}

	log.Println("Getting compiled program")
	compileErr := getCompiledProgram(err, generatedCode, tmpDir)
	if compileErr != nil {
		return compileErr
	}

	buildErr := prepareDockerFile(err, tmpDir)
	if buildErr != nil {
		return buildErr
	}

	log.Println("Building image")
	errBuild := BuildImageInTmpFolder(tmpDir, promptID)
	if errBuild != nil {
		return errBuild
	}
	return nil
}

func prepareDockerFile(err error, tmpDir string) error {
	// Create minimal Dockerfile
	dockerfile := `FROM alpine:3.21
WORKDIR /app
COPY main .
CMD ["./main"]`

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
	if err != nil {
		return err
	}
	return nil
}

func getCompiledProgram(err error, generatedCode string, tmpDir string) error {
	// Compile program using remote service
	compilerURL := os.Getenv("COMPILER_URL")
	if compilerURL == "" {
		return fmt.Errorf("no compiler URL set")
	}
	login := os.Getenv("COMPILER_LOGIN")
	password := os.Getenv("COMPILER_PASSWORD")
	req, err := http.NewRequest("POST", compilerURL, strings.NewReader(generatedCode))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(login, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("failed to compile: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("compilation failed: %s", string(body))
	}

	// Save compiled binary
	binaryPath := filepath.Join(tmpDir, "main")
	binaryFile, err := os.Create(binaryPath)
	if err != nil {
		return err
	}
	defer binaryFile.Close()

	written, err := io.Copy(binaryFile, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("written bytes: ", written)

	// Make binary executable
	return os.Chmod(binaryPath, 0755)
}

func AddGeneratedCodeToTheGameTemplate(txt string) (string, error) {
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
