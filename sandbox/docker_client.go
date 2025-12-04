package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Using a Facade Pattern ((hiding complexity behind a clean interface))
// Client is a wrapper around the official Docker Client
type Client struct {
	api *client.Client
}

// Next Block is very Important this is the main HANDSHAKING PART
func NewClient() (*Client, error) {
	// "FromEnv" is the MVP here.
	// It tells Go: "Check the environment variables."
	// On Windows, it finds the Named Pipe (npipe:////./pipe/docker_engine).
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// "WithAPIVersionNegotiation" prevents errors if your Client is newer than the Docker Engine.
	if err != nil {
		return nil, fmt.Errorf("Failed to create docker client: %w", err)
	}
	return &Client{api: cli}, nil
}

// Ping Checks if the daemon is actually reachable
// We should never assume a connection works always verify it
// Basically Sending a packet
func (c *Client) Ping(ctx context.Context) error {
	//context.Context controls how long we wait, because cant be waiting forever
	_, err := c.api.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker Daemon is unreacable: %w", err)
	}
	return nil
}

// Close the connection after using otherwise memory leaks
func (c *Client) Close() error {
	return c.api.Close()
}

// EnsureImage function if it exisits locally
func (c *Client) EnsureImage(ctx context.Context, imageName string) error {
	_, _, err := c.api.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return fmt.Errorf("Docker image '%s' not found. Did you run 'Docker Build'?", imageName)
		}
		return fmt.Errorf("Failer to check Docker Image: %w", err)
	}
	return nil
}

// RunContainer creates a container runs a command and returns the output
func (c *Client) RunContainer(ctx context.Context, image string, cmd []string) ([]byte, error) {
	config := &container.Config{
		Image: image, //Which Os
		Cmd:   cmd,   //What to do?
		Tty:   false, // False = clean Text Output
	}
	//Create the container (The Constrution)
	//We pass 'NIL' for host config because we arent settling Limits.....YET
	resp, err := c.api.ContainerCreate(ctx, config, nil, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	//Safety Switch[Deadman's Switch]
	//Schedule Cleanup
	defer func() {
		//Force : True [Means I dont care what happens KILL IT]
		removeOption := container.RemoveOptions{Force: true}
		c.api.ContainerRemove(context.Background(), containerID, removeOption)
	}()
	if err := c.api.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	//Starting the container
	//Race Condition (Wait vs Timeout)
	//Two Channels
	statusCh, errCh := c.api.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		//If Docker crashes
		if err != nil {
			return nil, fmt.Errorf("Error while waiting for container: %w", err)
		}
	case <-statusCh:
		//Successss
	case <-ctx.Done():
		//Timer ran out
		//So now the Defer function will run and kill
		return nil, fmt.Errorf("Execution timed out")
	}

	//Getting the logs now
	out, err := c.api.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch logs: %w", err)
	}
	//StdCopy splits the raw stream into Stdout and StdErr
	//For now we merge em together into one buffer to keep it simple
	var outputBuffer bytes.Buffer
	_, err = stdcopy.StdCopy(&outputBuffer, &outputBuffer, out)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}
	return outputBuffer.Bytes(), nil

}
