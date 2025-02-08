package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	maxUploadSize  = 10 << 20 // 10 MB
	compileTimeout = 30 * time.Second
)

func main() {
	if os.Getenv("AUTH_USERNAME") == "" || os.Getenv("AUTH_PASSWORD") == "" {
		fmt.Println("AUTH_USERNAME and AUTH_PASSWORD environment variables must be set")
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/", handleCompile)

	fmt.Println("Server starting on port 8080...")
	server.ListenAndServe()
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleCompile(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if username != os.Getenv("AUTH_USERNAME") || password != os.Getenv("AUTH_PASSWORD") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	sourceCode, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	tmpfile, err := os.CreateTemp("", "main*.go")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(sourceCode); err != nil {
		http.Error(w, "Error writing to temporary file", http.StatusInternalServerError)
		return
	}
	tmpfile.Close()

	outputFile := tmpfile.Name() + ".out"
	cmd := exec.Command("go", "build", "-o", outputFile, tmpfile.Name())
	stdErrBuf := new(strings.Builder)
	cmd.Stderr = stdErrBuf
	stdOutBuf := new(strings.Builder)
	cmd.Stdout = stdOutBuf

	// Add timeout
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("Compilation error: " + err.Error())
			fmt.Println("std err", stdErrBuf.String())
			fmt.Println("std out", stdOutBuf.String())
			http.Error(
				w,
				"Compilation error: "+err.Error()+stdErrBuf.String()+stdOutBuf.String(),
				http.StatusBadRequest,
			)
			return
		}
	case <-time.After(compileTimeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		http.Error(w, "Compilation timeout", http.StatusRequestTimeout)
		return
	}

	defer os.Remove(outputFile)
	compiledProgram, err := os.ReadFile(outputFile)
	if err != nil {
		http.Error(w, "Error reading compiled program", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=main")
	w.Write(compiledProgram)
}
