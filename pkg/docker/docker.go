// pkg/docker/docker.go
package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Cdaprod/repocate/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// DockerClient defines the abstraction for Docker operations.
type DockerClient interface {
	BuildImage(dockerfilePath, tag string) error
	CreateContainer(config container.Config, hostConfig container.HostConfig, networkConfig *network.NetworkingConfig) (string, error)
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error
	ListContainers() ([]types.Container, error)
	CreateNetwork(networkName string) error
	RemoveNetwork(networkName string) error
	ConnectContainerToNetwork(containerID, networkName string) error
}

// dockerClient is the concrete implementation of DockerClient using Docker SDK.
type dockerClient struct {
	cli    *client.Client
	logger *utils.Logger
}

// NewDockerClient creates a new instance of DockerClient.
func NewDockerClient(logger *utils.Logger) DockerClient {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to create Docker client: %v", err))
		os.Exit(1)
	}
	return &dockerClient{
		cli:    cli,
		logger: logger,
	}
}

// BuildImage builds a Docker image from a Dockerfile.
func (dc *dockerClient) BuildImage(dockerfilePath, tag string) error {
	dc.logger.LogInfo(fmt.Sprintf("Building Docker image from %s with tag %s...", dockerfilePath, tag))
	ctx := context.Background()

	dockerfile, err := os.Open(dockerfilePath)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to open Dockerfile: %v", err))
		return err
	}
	defer dockerfile.Close()

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: filepath.Base(dockerfilePath),
		Remove:     true,
		// Add more build options as needed
	}

	response, err := dc.cli.ImageBuild(ctx, dockerfile, buildOptions)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to build Docker image: %v", err))
		return err
	}
	defer response.Body.Close()

	// Stream build output to stdout or log
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to stream Docker build output: %v", err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker image %s built successfully.", tag))
	return nil
}

// CreateContainer creates a new Docker container with the specified configurations.
func (dc *dockerClient) CreateContainer(config container.Config, hostConfig container.HostConfig, networkConfig *network.NetworkingConfig) (string, error) {
	dc.logger.LogInfo(fmt.Sprintf("Creating Docker container from image %s...", config.Image))
	ctx := context.Background()

	resp, err := dc.cli.ContainerCreate(ctx, &config, &hostConfig, networkConfig, nil, "")
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to create Docker container: %v", err))
		return "", err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker container created with ID: %s", resp.ID))
	return resp.ID, nil
}

// StartContainer starts an existing Docker container.
func (dc *dockerClient) StartContainer(containerID string) error {
	dc.logger.LogInfo(fmt.Sprintf("Starting Docker container %s...", containerID))
	ctx := context.Background()

	err := dc.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to start Docker container %s: %v", containerID, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker container %s started successfully.", containerID))
	return nil
}

// StopContainer stops a running Docker container.
func (dc *dockerClient) StopContainer(containerID string) error {
	dc.logger.LogInfo(fmt.Sprintf("Stopping Docker container %s...", containerID))
	ctx := context.Background()

	timeout := 10 // seconds
	err := dc.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to stop Docker container %s: %v", containerID, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker container %s stopped successfully.", containerID))
	return nil
}

// RemoveContainer removes a Docker container.
func (dc *dockerClient) RemoveContainer(containerID string) error {
	dc.logger.LogInfo(fmt.Sprintf("Removing Docker container %s...", containerID))
	ctx := context.Background()

	err := dc.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to remove Docker container %s: %v", containerID, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker container %s removed successfully.", containerID))
	return nil
}

// ListContainers lists all Docker containers.
func (dc *dockerClient) ListContainers() ([]types.Container, error) {
	dc.logger.LogInfo("Listing Docker containers...")
	ctx := context.Background()

	containers, err := dc.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to list Docker containers: %v", err))
		return nil, err
	}

	dc.logger.LogInfo("Docker containers listed successfully.")
	return containers, nil
}

// CreateNetwork creates a new Docker network if it doesn't exist.
func (dc *dockerClient) CreateNetwork(networkName string) error {
	dc.logger.LogInfo(fmt.Sprintf("Creating Docker network '%s'...", networkName))
	ctx := context.Background()

	// Check if network already exists
	networks, err := dc.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to list Docker networks: %v", err))
		return err
	}

	for _, net := range networks {
		if net.Name == networkName {
			dc.logger.LogInfo(fmt.Sprintf("Docker network '%s' already exists.", networkName))
			return nil
		}
	}

	// Create network
	_, err = dc.cli.NetworkCreate(ctx, networkName, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to create Docker network '%s': %v", networkName, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker network '%s' created successfully.", networkName))
	return nil
}

// RemoveNetwork removes a Docker network.
func (dc *dockerClient) RemoveNetwork(networkName string) error {
	dc.logger.LogInfo(fmt.Sprintf("Removing Docker network '%s'...", networkName))
	ctx := context.Background()

	err := dc.cli.NetworkRemove(ctx, networkName)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to remove Docker network '%s': %v", networkName, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker network '%s' removed successfully.", networkName))
	return nil
}

// ConnectContainerToNetwork connects a Docker container to a specified network.
func (dc *dockerClient) ConnectContainerToNetwork(containerID, networkName string) error {
	dc.logger.LogInfo(fmt.Sprintf("Connecting Docker container %s to network '%s'...", containerID, networkName))
	ctx := context.Background()

	err := dc.cli.NetworkConnect(ctx, networkName, containerID, nil)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to connect Docker container %s to network '%s': %v", containerID, networkName, err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Docker container %s connected to network '%s' successfully.", containerID, networkName))
	return nil
}

// ExecCommand executes a command inside a running Docker container.
// It streams the output to the host's stdout and stderr.
func (dc *dockerClient) ExecCommand(containerID string, cmd []string) error {
	dc.logger.LogInfo(fmt.Sprintf("Executing command '%v' in container %s...", cmd, containerID))
	ctx := context.Background()

	// Create Exec instance
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}
	execIDResp, err := dc.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to create exec instance: %v", err))
		return err
	}

	// Attach to the Exec instance
	resp, err := dc.cli.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{})
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to attach to exec instance: %v", err))
		return err
	}
	defer resp.Close()

	// Stream the output
	output, err := dc.readExecOutput(resp.Reader)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to read exec output: %v", err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Command Output:\n%s", output))
	return nil
}

// readExecOutput reads the output from the exec command and returns it as a string.
func (dc *dockerClient) readExecOutput(reader io.Reader) (string, error) {
	var output string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return output, nil
}

// TagImage tags an existing Docker image with a new tag.
// Useful for versioning or preparing images for pushing to a registry.
func (dc *dockerClient) TagImage(sourceImage, targetImage string) error {
	dc.logger.LogInfo(fmt.Sprintf("Tagging image '%s' as '%s'...", sourceImage, targetImage))
	ctx := context.Background()

	err := dc.cli.ImageTag(ctx, sourceImage, targetImage)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to tag image: %v", err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Image '%s' tagged as '%s' successfully.", sourceImage, targetImage))
	return nil
}

// PushImage pushes a Docker image to a specified Docker registry.
// Requires authentication configuration.
func (dc *dockerClient) PushImage(image string, authConfig types.AuthConfig) error {
	dc.logger.LogInfo(fmt.Sprintf("Pushing image '%s' to registry...", image))
	ctx := context.Background()

	// Encode the authentication credentials
	encodedAuth, err := client.EncodeAuthToBase64(authConfig)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to encode auth config: %v", err))
		return err
	}

	pushOptions := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	}

	response, err := dc.cli.ImagePush(ctx, image, pushOptions)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to push image: %v", err))
		return err
	}
	defer response.Close()

	// Stream push output to stdout
	_, err = io.Copy(os.Stdout, response)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to stream push output: %v", err))
		return err
	}

	dc.logger.LogInfo(fmt.Sprintf("Image '%s' pushed to registry successfully.", image))
	return nil
}

// RebuildContainer rebuilds the Docker image and recreates the container with the new image.
// It stops and removes the existing container, builds the new image, and starts a new container.
func (dc *dockerClient) RebuildContainer(containerID string, newImage string) error {
	dc.logger.LogInfo(fmt.Sprintf("Rebuilding container %s with image '%s'...", containerID, newImage))
	ctx := context.Background()

	// Inspect the existing container to retrieve its configuration
	containerJSON, err := dc.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to inspect container %s: %v", containerID, err))
		return err
	}

	// Stop the existing container
	err = dc.StopContainer(containerID)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to stop container %s: %v", containerID, err))
		return err
	}

	// Remove the existing container
	err = dc.RemoveContainer(containerID)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to remove container %s: %v", containerID, err))
		return err
	}

	// Tag the new image appropriately (optional)
	// Example: "user/repo:latest"
	// err = dc.TagImage(newImage, "user/repo:latest")
	// if err != nil {
	// 	return err
	// }

	// Recreate the container with the new image
	newContainerID, err := dc.CreateContainer(containerJSON.Config, containerJSON.HostConfig, containerJSON.NetworkSettings.Networks)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to create new container with image '%s': %v", newImage, err))
		return err
	}

	// Start the new container
	err = dc.StartContainer(newContainerID)
	if err != nil {
		dc.logger.LogError(fmt.Sprintf("Failed to start new container %s: %v", newContainerID, err))
		return err
	}

	// Connect the new container to the network(s)
	for networkName := range containerJSON.NetworkSettings.Networks {
		err = dc.ConnectContainerToNetwork(newContainerID, networkName)
		if err != nil {
			dc.logger.LogError(fmt.Sprintf("Failed to connect container %s to network '%s': %v", newContainerID, networkName, err))
			return err
		}
	}

	dc.logger.LogInfo(fmt.Sprintf("Container %s rebuilt and started successfully with image '%s'.", newContainerID, newImage))
	return nil
}

// -------------------- End of New Functions --------------------