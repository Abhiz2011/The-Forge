package main

import (
	"Forge/api"
	"Forge/sandbox"
	"context"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔥 The Forge is starting up...")
	//1. Intialzing the Docker Engine
	fmt.Println("1. Connecting to the Docker Engine...")
	dockerClient, err := sandbox.NewClient()
	if err != nil {
		panic(err) // IF docker down , I cannot run this
	}
	defer dockerClient.Close()
	//2. Performing a health check on Docker
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dockerClient.Ping(ctx); err != nil {
		panic(err)
	}
	fmt.Println("   >> Docker Connected! 🐳")
	//3. Ensure the image Exists
	if err := dockerClient.EnsureImage(ctx, "forge-cpp-runner"); err != nil {
		panic(err)
	}
	fmt.Println("   >> Sandbox Image Verified! ✅")

	//4. Dependency Injection
	forgeAPI := &api.API{
		Sandbox: dockerClient,
	}
	//5. Starting the HTTP Server
	fmt.Println("2. Starting API Server on port :3000...")

	//Now using the methods from our new ForgeAPI Struct
	http.HandleFunc("/health", forgeAPI.HealthHandler)
	http.HandleFunc("/submit", forgeAPI.SubmitHandler)
	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}
