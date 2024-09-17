package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Get the current working directory
func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

// Open the default browser depending on the OS
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start() // Linux
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() // Windows
	case "darwin":
		err = exec.Command("open", url).Start() // macOS
	default:
		fmt.Println("Unsupported platform. Please open the browser manually.")
		return
	}

	if err != nil {
		log.Printf("Failed to open the browser: %v", err)
	}
}

func startGodoc() {
	err := exec.Command("godoc", "-http", ":6060").Start()
	if err != nil {
		log.Printf("Failed to execute the godoc command: %v", err)
	}
}

// Serve static files from the specified directory
// This is the entry point of the program
// Get the current directory where the HTML documentation is located
func main() {
	startGodoc()
	dir := getWorkingDir()

	documentation := filepath.Join(dir, "template", "docs")

	// Serve static files from the current directory
	http.Handle("/", http.FileServer(http.Dir(documentation)))

	port := "8085"
	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("Serving documentation at %s\n", url)

	openBrowser(url)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
