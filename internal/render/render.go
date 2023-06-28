package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/justinas/nosurf"
	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/models"
)

var app *config.AppConfig
var pathToTemplate = "./templates"

func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	return td
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error{
	var templateCache map[string]*template.Template
	if app.UseCache {
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	requestedTemplate, ok := templateCache[tmpl]
	if !ok {
		log.Println("Could not get template from cache!")
		return errors.New("cannot get template from cache")
	}

	buf := new(bytes.Buffer)
	td := AddDefaultData(templateData, r)
	err := requestedTemplate.Execute(buf, td)

	if err != nil {
		log.Println(err)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	// templateCache := make(map[string]*template.Template)
	templateCache := map[string]*template.Template{}

	//get all template files
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return templateCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err != nil {
			return templateCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err != nil {
				return templateCache, err
			}
		}
		templateCache[name] = ts
	}
	return templateCache, nil
}
