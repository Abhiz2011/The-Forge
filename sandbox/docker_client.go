package sandbox

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
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
func (c *Client) RunContainer(ctx context.Context, image string, code string) ([]byte, error) {
	//We are HARDCODING the command to compile and run
	//"sh-c" lets us chain commands using &&
	cmd := []string{"sh", "-c", "g++ main.cpp -o main && ./main"}
	config := &container.Config{
		Image: image, //Which Os
		Cmd:   cmd,   //What to do?
		Tty:   false, // False = clean Text Output
	}
	//Create the container (The Constrution)
	//passed 'NIL' for host config because I am not setting Limits.....YET
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
	//Injecting the code into out "TELEPORTER"
	//Turning the code string into a TAR Archive (in RAM)
	tarReader, err := createTar(code)
	if err != nil {
		return nil, fmt.Errorf("Failed to create tar archive : %w", err)
	}
	//Now shipping it to Docker Container
	err = c.api.CopyToContainer(ctx, containerID, "/app/", tarReader, types.CopyToContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to copy code to container: %w", err)
	}
	//Starting the container
	if err := c.api.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
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

// Now CreateTar func to convert string of code into a filesystem
func createTar(code string) (io.Reader, error) {
	//(1) the in memory buffer
	var buf bytes.Buffer
	//(2) The writer helps us in complexity to format data as a TAR file
	tw := tar.NewWriter(&buf)

	//(3) The Header(Otherwise known as The Envelope)
	header := &tar.Header{
		Name: "main.cpp",
		Mode: 0777, //Permissions (Read/Write/Execute for everyone)
		Size: int64(len(code)),
	}
	//Write the header first
	if err := tw.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("Failed to write the TAR Header: %w", err)
	}
	//(4) The Body now we are writing the actual c++ code
	if _, err := tw.Write([]byte(code)); err != nil {
		return nil, fmt.Errorf("Failed to write the TAR Body: %w", err)
	}

	//(5) Close the writer
	//This adds end of file marker to the archive so that docker knows where it ends
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("Failed to close the writer: %w", err)
	}
	//Return the memory buffer which is now a TAR File
	return &buf, nil
}
