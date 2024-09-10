package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"vtube/services"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func processVideoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Respond immediately that processing has started
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"message": "Video processing started"}
		json.NewEncoder(w).Encode(response)

		// Start video processing in a separate goroutine
		go func() {
			// Hardcoded directories for simplicity
			inputDir := "./video/input"
			outputDir := "./video/output"
			err := services.ProcessVideo(inputDir, outputDir)
			if err != nil {
				fmt.Printf("Error processing videos: %v\n", err)
			}
		}()
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func main() {
	fmt.Println("hello world")
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Println("cpus ", numCPU)

	fs := http.FileServer(http.Dir("./video/output"))
	http.Handle("/vtube/video/output/", enableCORS(http.StripPrefix("/vtube/video/output/", fs)))

	http.HandleFunc("/vtube/process", processVideoHandler)

	port := ":8080"
	fmt.Printf("Server started at http://localhost%s\n", port)

	// Start the server
	log.Fatal(http.ListenAndServe(port, nil))

	fmt.Println("video processing completed successfully")
}
