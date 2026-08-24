package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mutex    = &sync.Mutex{}
)

func generateShortCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]

	mutex.Lock()
	originalURL, exists := urlStore[shortCode]
	mutex.Unlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	var longURL string

	// Check if a URL was passed directly as a command-line argument
	if len(os.Args) > 1 {
		longURL = os.Args[1]
	} else {
		// Otherwise, prompt the user to type it in the terminal
		fmt.Print("Enter the long URL to shorten: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		longURL = strings.TrimSpace(input)
	}

	if longURL == "" {
		fmt.Println("Error: No URL provided.")
		return
	}

	// Store the URL mapping in memory
	mutex.Lock()
	shortCode := generateShortCode()
	urlStore[shortCode] = longURL
	mutex.Unlock()

	// Start the redirect server in a background goroutine
	http.HandleFunc("/", redirectHandler)
	go func() {
		if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	// String concatenation: Base domain + generated code
	baseURL := "http://127.0.0.1:8080/"
	shortenedURL := baseURL + shortCode

	fmt.Println("\n----------------------------------------")
	fmt.Println("Original URL: ", longURL)
	fmt.Println("Shortened URL:", shortenedURL)
	fmt.Println("----------------------------------------")
	fmt.Println("Server is running. Open the short link in your browser!")
	fmt.Println("Press Ctrl+C to exit.")

	// Block the main thread so the process and server stay alive
	select {}
}
