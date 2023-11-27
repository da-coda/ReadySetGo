package services

import (
	"ReadySetGo/util"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type ProjectService interface {
	CreateNewProject(originalFile io.Reader, name string) (string, error)
	LoadProject(projectSlug string) (ProjectConfig, error)
	GetProjects() ([]ProjectConfig, error)
}

type projectService struct {
	binBaseDir    string
	dockerService DockerService
}

func NewProjectService(binBaseDir string, dockerService DockerService) ProjectService {
	return projectService{binBaseDir: binBaseDir, dockerService: dockerService}
}

func (s projectService) CreateNewProject(originalFile io.Reader, projectName string) (string, error) {
	sanitizedProjectName := strings.Replace(projectName, " ", "_", -1)
	if sanitizedProjectName == "" {
		return "", fmt.Errorf("no project name provided")
	}
	var forbiddenChars = regexp.MustCompile(`[^a-zA-Z0-9_ ]+`)
	sanitizedProjectName = forbiddenChars.ReplaceAllString(sanitizedProjectName, "")
	sanitizedProjectName = strings.ToLower(sanitizedProjectName)
	dirPath := filepath.Clean(fmt.Sprintf("%s/%s", s.binBaseDir, sanitizedProjectName))
	err := os.Mkdir(dirPath, 0700)
	if err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}
	dst, err := os.Create(fmt.Sprintf("%s/executable", dirPath))
	if err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, originalFile); err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}
	isStaticallyLinked, err := util.IsStaticallyLinkedBinary(dst.Name())
	if err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}

	if !isStaticallyLinked {
		return "", fmt.Errorf("uploaded file must be a statically linked binary")
	}
	if err := createProjectConfig(dirPath, projectName, sanitizedProjectName); err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}
	err = s.dockerService.InitDockerProject(dirPath, sanitizedProjectName)
	if err != nil {
		return "", fmt.Errorf("unable to create project: %w", err)
	}
	return sanitizedProjectName, nil
}

func (s projectService) LoadProject(projectSlug string) (ProjectConfig, error) {
	dirPath := filepath.Clean(fmt.Sprintf("%s/%s", s.binBaseDir, projectSlug))
	isDirectory, err := util.IsDirectory(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load project: %w", err)
	}
	if !isDirectory {
		return nil, fmt.Errorf("could not find project directory")
	}
	return loadConfig(dirPath)
}

func (s projectService) GetProjects() ([]ProjectConfig, error) {
	var projects []ProjectConfig
	dirEntries, err := os.ReadDir(s.binBaseDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read bin directory: %w", err)
	}
	mu := &sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, entry := range dirEntries {
		wg.Add(1)
		go func(dirEntry os.DirEntry) {
			defer wg.Done()
			config, err := loadConfig(fmt.Sprintf("%s/%s", s.binBaseDir, dirEntry.Name()))
			if err != nil {
				slog.Error(fmt.Errorf("unable to load config for project %s: %w", dirEntry.Name(), err).Error())
				return
			}
			mu.Lock()
			projects = append(projects, config)
			mu.Unlock()
		}(entry)
	}
	wg.Wait()
	return projects, nil
}

type ProjectConfig interface {
	GetVersion() string
	GetName() string
	GetPorts() []int
	GetEnvs() map[string]string
	GetSlug() string
}

type projectConfigV1 struct {
	Version string            `toml:"version"`
	Slug    string            `toml:"slug"`
	Name    string            `toml:"name"`
	Ports   []int             `toml:"ports"`
	Envs    map[string]string `toml:"envs"`
}

func createProjectConfig(projectDir string, name string, slug string) error {
	config := projectConfigV1{Version: "v1", Name: name, Slug: slug, Ports: make([]int, 0), Envs: make(map[string]string)}
	configFile, err := os.Create(fmt.Sprintf("%s/config", projectDir))
	if err != nil {
		return fmt.Errorf("unable to write config: %w", err)
	}
	encoder := toml.NewEncoder(configFile)
	return encoder.Encode(config)
}

func loadConfig(projectDir string) (ProjectConfig, error) {
	file, err := os.Open(fmt.Sprintf("%s/config", projectDir))
	if err != nil {
		return nil, err
	}
	version, err := parseConfigVersion(file)
	if err != nil {
		return nil, err
	}
	switch version {
	case "v1":
		return loadV1Config(file)
	default:
		return nil, fmt.Errorf("unknown config version %s", version)
	}
}

func loadV1Config(file *os.File) (ProjectConfig, error) {
	config := &projectConfigV1{}
	decoder := toml.NewDecoder(file)
	_, err := decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}
	return config, nil
}

func parseConfigVersion(file *os.File) (string, error) {
	configVersion := &struct {
		Version string `toml:"version"`
	}{}
	decoder := toml.NewDecoder(file)
	_, err := decoder.Decode(configVersion)
	if err != nil {
		return "", fmt.Errorf("unable to parse config version: %w", err)
	}
	file.Seek(0, 0)
	return configVersion.Version, nil
}

func (p projectConfigV1) GetVersion() string {
	return p.Version
}

func (p projectConfigV1) GetName() string {
	return p.Name
}

func (p projectConfigV1) GetPorts() []int {
	return p.Ports
}

func (p projectConfigV1) GetEnvs() map[string]string {
	return p.Envs
}

func (p projectConfigV1) GetSlug() string {
	return p.Slug
}
