package main

//go:generate bin/templ generate
import (
	"ReadySetGo/config"
	_ "ReadySetGo/config"
	"ReadySetGo/handler"
	"ReadySetGo/services"
	"ReadySetGo/templates/pages"
	"embed"
	"fmt"
	"github.com/a-h/templ"
	dockerClient "github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

var (
	//go:embed assets
	assets embed.FS
)

func main() {
	client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err.Error())
	}
	dockerService := services.NewDockerService(client)
	projectService := services.NewProjectService(config.GetBinPath(), dockerService)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", templ.Handler(pages.Index()).ServeHTTP)
	r.Route("/create", func(r chi.Router) {
		r.Get("/", templ.Handler(pages.CreateNew()).ServeHTTP)
		r.Post("/", handler.CreateUploadHandler(projectService))
	})
	r.Route("/projects", func(r chi.Router) {
		r.Get("/{projectSlug}", handler.CreateProjectGetHandler(projectService))
		r.Get("/{projectSlug}/control", handler.CreateProjectGetControlTabHandler())
		r.Post("/{projectSlug}/control/start", handler.CreateProjectStartHandler(dockerService))
	})
	r.Handle("/*", http.FileServer(http.FS(assets)))
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.GetPort()), r)
	if err != nil {
		panic(err.Error())
	}
}
