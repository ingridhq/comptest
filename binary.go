package comptest

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

// RunBinary will run golang app in a background. Returns clean function
func RunBinary(pathToBinary string, pathToLogs string) (func(), error) {
	cmd := exec.Command(pathToBinary)
	outfile, err := os.Create(pathToLogs)
	if err != nil {
		return nil, fmt.Errorf("couldn't create file: %v", err)
	}
	cmd.Stdout, cmd.Stderr = outfile, outfile
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("couldn't start grpc server: %v", err)
	}

	// closed indicates if "we" called closed function. Concurrent access is synchronized using closedMtx.
	closed := false
	closedMtx := &sync.Mutex{}

	go func() {
		// Wait for the command to exit. If it is "us" who closed (killed) the command, do nothing.
		// If the command exited spontaneously, there is no point to continue the tests.
		err := cmd.Wait()
		closedMtx.Lock()
		defer closedMtx.Unlock()
		if !closed {
			// Print logs error to stdout, to make debugging on CircleCI easier.
			bytes, readErr := os.ReadFile(pathToLogs)
			if readErr != nil {
				log.Printf("could not read log file.\n")
			} else {
				log.Printf("%s\n", bytes)
			}

			// At this point we don't have access to *testing.T.
			// Therefore the only thing we can do is panic using log.Fatal.
			log.Fatalf("child process %q exited before it was closed by tests: %v", pathToBinary, err)
		}
	}()

	return func() {
		closedMtx.Lock()
		defer closedMtx.Unlock()

		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill sut process: %v", err)
		}
		if err := outfile.Close(); err != nil {
			log.Printf("Failed to close logs file: %v", err)
		}
		closed = true
	}, nil
}

func MustRunBinary(pathToBinary string, pathToLogs string) func() {
	clean, err := RunBinary(pathToBinary, pathToLogs)
	if err != nil {
		log.Fatal(err)
	}
	return clean
}

// BuildBinary will build golang application.
func BuildBinary(pathToGoMain string, pathToBinary string) error {
	cmd := exec.Command("go", "build", "-o", pathToBinary, pathToGoMain)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't build go app: %v", err)
	}
	return nil
}

func MustBuildBinary(pathToGoMain string, pathToBinary string) {
	if err := BuildBinary(pathToGoMain, pathToBinary); err != nil {
		log.Fatal(err)
	}
}

// MustBuildAndRun will build and run golang app in a background. Returns clean function
func MustBuildAndRun(pathToGoMain, pathToBinary, pathToLogs string) func() {
	MustBuildBinary(pathToGoMain, pathToBinary)
	return MustRunBinary(pathToBinary, pathToLogs)
}
