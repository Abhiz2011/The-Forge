package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

type Submission struct {
	Code string `json:"code"`
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		//Send 405 NOT ALLOWED
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//Create a variable of my struct
	var sub = Submission{}                      // Easy way to understand is to Create a Bucket
	err := json.NewDecoder(r.Body).Decode(&sub) // Attach a hose for controlled release of water into the bucket
	//err := decoder.Decode(&sub)    // Fill the bucket, using & pass by ref fills the memory of the bucket
	if err != nil {
		http.Error(w, "Bad Json", http.StatusBadRequest)
		return
	}

	fmt.Println("Received Code, Processing")
	result := processSubmission(sub.Code)
	fmt.Fprintf(w, result)
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Check Received")
	fmt.Fprintf(w, "System Operational")
}

func main() {
	fmt.Println("Starting on :3000...")
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/submit", submitHandler)
	http.ListenAndServe(":3000", nil)

}

func processSubmission(code string) string {
	//Write File
	err := os.WriteFile("submission.cpp", []byte(code), 0644)
	if err != nil {
		return "Failed to Write" + err.Error()
	}
	//Compile File
	cmd := exec.Command("g++", "submission.cpp", "-o", "submission.exe")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Compilation Failed!:\n" + string(output)
	}
	//Run the file
	runCmd := exec.Command("./submission.exe")
	runOutput, runErr := runCmd.CombinedOutput()
	if runErr != nil {
		return "Runtime Error: \n" + string(runOutput)
	}
	return string(runOutput)
}
