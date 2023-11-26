package handler

import (
	"ReadySetGo/services"
	"ReadySetGo/templates/components"
	"fmt"
	"github.com/a-h/templ"
	"net/http"
)

func CreateUploadHandler(service services.ProjectService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(50 << 20) // 10 MB
		if err != nil {
			templ.Handler(components.UploadError(err.Error())).ServeHTTP(w, r)
			return
		}
		projectName := r.FormValue("name")
		file, _, err := r.FormFile("file")
		if err != nil {
			templ.Handler(components.UploadError(err.Error())).ServeHTTP(w, r)
			return
		}
		if err != nil {
			templ.Handler(components.UploadError(err.Error())).ServeHTTP(w, r)
			return
		}

		defer file.Close()
		projectSlug, err := service.CreateNewProject(file, projectName)
		if err != nil {
			templ.Handler(components.UploadError(err.Error())).ServeHTTP(w, r)
			return
		}
		redirectURI := fmt.Sprintf("/projects/%s", projectSlug)
		w.Header().Add("HX-Redirect", redirectURI)
		w.WriteHeader(201)
	}
}
