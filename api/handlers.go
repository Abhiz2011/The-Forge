package api

import (
	"Forge/sandbox"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Api holds dependancies that our handlers need
// we "inject" the sandbox here so that our handlers can use it
type API struct {
	Sandbox *sandbox.Client
}

// Submission is the JSON shape i need
type Submission struct {
	Code string `json:"code"`
}

// SubmitHandler handles the POST/ submit request
func (a *API) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	//1. Method Guard
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	//2. Parse JSON
	var sub Submission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid JSON Format", http.StatusBadRequest)
		return
	}

	//3. Validation
	if len(sub.Code) == 0 {
		http.Error(w, "Code cannot be empty", http.StatusBadRequest)
		return
	}

	//4. Integration
	fmt.Println("Received code. Sending to Sandbox...")
	//Creating a context with a timeout of 10 sec
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Call the Engine
	output, err := a.Sandbox.RunContainer(ctx, "forge-cpp-runner", sub.Code)
	if err != nil {
		//Log Internal error
		fmt.Printf("Sandbox Error: %v\n", err)
		//Send generic error to user
		http.Error(w, fmt.Sprintf("Execution Failed: %v", err), http.StatusInternalServerError)
		return
	}
	//5. Success Response
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}
func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The Forge is active "))
}
