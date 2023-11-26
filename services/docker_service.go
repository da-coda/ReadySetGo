package services

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"io"
	"os"
	"time"
)

type DockerService interface {
	InitDockerProject(projectPath string, projectSlug string) error
	Dockerfile() []byte
	RunProject(projectSlug string) error
}

type dockerService struct {
	dockerClient *client.Client
}

func NewDockerService(dockerClient *client.Client) DockerService {
	return dockerService{dockerClient: dockerClient}
}

func (d dockerService) InitDockerProject(projectPath string, projectSlug string) error {
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", projectPath)
	dockerfile, err := os.Create(dockerfilePath)
	if err != nil {
		return fmt.Errorf("unable to create Dockerfile: %w", err)
	}
	defer dockerfile.Close()
	_, err = dockerfile.Write(d.Dockerfile())
	if err != nil {
		return fmt.Errorf("unable to create Dockerfile: %w", err)
	}
	err = d.imageBuild(projectPath, projectSlug)
	if err != nil {
		return fmt.Errorf("unable to build image: %w", err)
	}
	return nil
}

func (d dockerService) Dockerfile() []byte {
	return []byte(`FROM scratch
COPY --chmod=777 ./executable /go/executable
ENTRYPOINT ["/go/executable"]`)
}

func (d dockerService) RunProject(projectSlug string) error {
	resp, err := d.dockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: projectSlug,
		ExposedPorts: nat.PortSet{
			"3333/tcp": struct{}{},
		},
	},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"3333/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "3333",
					},
				},
			},
		}, nil, nil, projectSlug)
	if err != nil {
		return fmt.Errorf("unable to create docker container: %w", err)
	}

	if err := d.dockerClient.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("unable to start docker container: %w", err)
	}
	return nil
}

func (d dockerService) imageBuild(dockerDir, imageTag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	tar, err := archive.TarWithOptions(dockerDir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageTag},
		NoCache:    true,
		Version:    "2",
	}
	res, err := d.dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		return err
	}

	all, err := io.ReadAll(res.Body)
	fmt.Println(string(all))
	defer res.Body.Close()

	return nil
}
