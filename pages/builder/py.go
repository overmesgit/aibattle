package builder

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func GetPyProgram(generatedCode string, tmpDir string) error {
	log.Println("Getting compiled program")
	compileErr := savePyProgram(generatedCode, tmpDir)
	if compileErr != nil {
		return compileErr
	}

	buildErr := preparePyDockerFile(tmpDir)
	if buildErr != nil {
		return buildErr
	}
	return nil
}

func preparePyDockerFile(tmpDir string) error {
	// Create minimal Dockerfile
	dockerfile := `FROM python:3-alpine
WORKDIR /app
COPY main.py .
CMD ["python3", "main.py"]`

	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
	if err != nil {
		return err
	}
	return nil
}

func savePyProgram(generatedCode string, tmpDir string) error {
	// Save compiled binary
	binaryPath := filepath.Join(tmpDir, "main.py")
	binaryFile, err := os.Create(binaryPath)
	if err != nil {
		return err
	}
	defer binaryFile.Close()

	written, err := io.Copy(binaryFile, bytes.NewBufferString(generatedCode))
	if err != nil {
		return err
	}
	fmt.Println("written bytes: ", written)

	// Make binary executable
	return os.Chmod(binaryPath, 0755)
}
