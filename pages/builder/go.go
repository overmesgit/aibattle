package builder

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetGoProgram(generatedCode string, tmpDir string) error {
	log.Println("Getting compiled program")
	compileErr := getCompiledProgram(generatedCode, tmpDir)
	if compileErr != nil {
		return compileErr
	}

	buildErr := prepareDockerFile(tmpDir)
	if buildErr != nil {
		return buildErr
	}
	return nil
}

func prepareDockerFile(tmpDir string) error {
	// Create minimal Dockerfile
	dockerfile := `FROM alpine:3.21
WORKDIR /app
COPY main .
CMD ["./main"]`

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
	if err != nil {
		return err
	}
	return nil
}

func getCompiledProgram(generatedCode string, tmpDir string) error {
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

	if err != nil {
		return fmt.Errorf("failed to compile: %v", err)
	}
	defer resp.Body.Close()

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
