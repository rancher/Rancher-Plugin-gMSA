package magetools

import (
	"bufio"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	ctx "context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Docker struct {
	ctx context.Context
	cli *client.Client
}

type DockerLog struct {
	Msg   string `json:"stream"`
	Error string `json:"error"`
}

func NewDocker() *Docker {
	dockerCtx := ctx.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Docker{
		ctx: dockerCtx,
		cli: cli,
	}
}

func (d *Docker) Build(args map[string]*string, dockerFile string, tag string, context string, noCache bool) error {
	wd, _ := os.Getwd()
	if context != "." {
		wd = filepath.Join(wd, context)
	}

	buildCtx, _ := archive.TarWithOptions(wd, &archive.TarOptions{})

	resp, err := d.cli.ImageBuild(d.ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: dockerFile,
		BuildArgs:  args,
		Tags:       []string{tag},
		NoCache:    noCache,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := readOutput(resp.Body); err != nil {
		return err
	}

	logrus.Infof("Built %s", tag)

	return nil
}

func (d *Docker) Copy(containerName, filePath, destination string) error {
	container, err := d.getContainerByName(containerName)
	if err != nil {
		return fmt.Errorf("could not get container %s by name: %v", containerName, err)
	}

	cpy, stat, err := d.cli.CopyFromContainer(d.ctx, container.ID, filePath)
	if err != nil {
		return fmt.Errorf("failed to copy file '%s' from container: %v", filePath, err)
	}

	if stat.Size == 0 {
		return fmt.Errorf("file %s not found in container %s", filePath, containerName)
	}

	bytes, err := io.ReadAll(cpy)
	if err != nil {
		return fmt.Errorf("failed to read container file %s: %v", filePath, err)
	}

	f, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("failed to open destination file %s: %v", destination, err)
	}

	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write to destination file %s: %v", destination, err)
	}

	return nil
}

func (d *Docker) Create(containerName, image string) (container.CreateResponse, error) {
	cnf := &container.Config{
		Image: image,
	}

	resp, err := d.cli.ContainerCreate(d.ctx, cnf, nil, nil, nil, containerName)
	if err != nil {
		return container.CreateResponse{}, fmt.Errorf("failed to create container %s from image %s: %v", containerName, image, err)
	}

	return resp, nil
}

func (d *Docker) Remove(containerName string) error {
	container, err := d.getContainerByName(containerName)
	if err != nil {
		return fmt.Errorf("could not get container %s by name: %v", containerName, err)
	}

	if err = d.cli.ContainerRemove(d.ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("could not remove container %s: %v", containerName, err)
	}

	return nil
}

func (d *Docker) getContainerByName(containerName string) (types.Container, error) {
	containers, err := d.cli.ContainerList(d.ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return types.Container{}, fmt.Errorf("failed to list containers: %v", err)
	}

	if len(containers) == 0 {
		return types.Container{}, fmt.Errorf("no container with name '%s' found", containerName)
	}

	for _, container := range containers {
		if len(container.Names) != 0 && container.Names[0] == "/"+containerName {
			return container, nil
		}
	}

	return types.Container{}, fmt.Errorf("no container with name '%s' found", containerName)
}

func (d *Docker) getContainerByImageTag(imageTag string) (types.ImageSummary, error) {
	containers, err := d.cli.ImageList(d.ctx, types.ImageListOptions{

		All: true,
	})
	if err != nil {
		return types.ImageSummary{}, fmt.Errorf("failed to list containers: %v", err)
	}

	if len(containers) == 0 {
		return types.ImageSummary{}, fmt.Errorf("no image with name '%s' found", imageTag)
	}

	for _, container := range containers {
		j, _ := json.MarshalIndent(container, "", " ")
		fmt.Println(string(j))

		if len(container.RepoTags) > 0 && container.RepoTags[0] == imageTag {
			return container, nil
		}
	}

	return types.ImageSummary{}, fmt.Errorf("no container with name '%s' found", imageTag)
}

func (d *Docker) runContainer(containerID string) error {
	err := d.cli.ContainerStart(d.ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container %s: %v", containerID, err)
	}

	containerStatusChan, errChan := d.cli.ContainerWait(d.ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err = <-errChan:
		if err != nil {
			return fmt.Errorf("error encountered waiting for container to finish running: %v", err)
		}
	case <-containerStatusChan:
	}

	out, err := d.cli.ContainerLogs(d.ctx, containerID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}

func (d *Docker) CreateManifest(manifestName string, images []string) error {

	var args []string
	args = append(args, "manifest")
	args = append(args, fmt.Sprintf("create"))
	args = append(args, manifestName)
	for _, image := range images {
		args = append(args, "--amend", image)
	}

	cmd := exec.Command("docker", args...)
	logrus.Infof("running command %s", cmd.String())
	// docker sdk doesn't have a good way to build manifests in go
	o, err := cmd.CombinedOutput()
	logrus.Infof(string(o))
	return err
}

func readOutput(body io.ReadCloser) error {
	scanner := bufio.NewScanner(bufio.NewReader(body))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		log := DockerLog{}
		b := scanner.Bytes()
		err := json.Unmarshal(b, &log)
		if err != nil {
			logrus.Error("failed to read docker log.")
		}

		if log.Msg != "" {
			fmt.Print(log.Msg)
		}

		if log.Error != "" {
			return fmt.Errorf(log.Error)
		}
	}
	return nil
}
