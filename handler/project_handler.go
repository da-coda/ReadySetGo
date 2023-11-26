package handler

import (
	"ReadySetGo/services"
	"ReadySetGo/templates/components"
	"ReadySetGo/templates/pages"
	"fmt"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func CreateProjectGetHandler(service services.ProjectService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectSlug := chi.URLParam(r, "projectSlug")
		projectConfig, err := service.LoadProject(projectSlug)
		if err != nil {
			templ.Handler(pages.Project(projectConfig, err)).ServeHTTP(w, r)
			return
		}
		templ.Handler(pages.Project(projectConfig, err)).ServeHTTP(w, r)
	}
}

func CreateProjectGetControlTabHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectSlug := chi.URLParam(r, "projectSlug")
		templ.Handler(components.ProjectControl(projectSlug)).ServeHTTP(w, r)
	}
}

func CreateProjectStartHandler(dockerService services.DockerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := dockerService.RunProject(chi.URLParam(r, "projectSlug"))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("<span>%s</span>", err.Error())))
			return
		}
		w.Write([]byte("<span>Success</span>"))
	}
}
