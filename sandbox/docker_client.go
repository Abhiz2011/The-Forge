package sandbox

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
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
