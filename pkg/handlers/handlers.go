package handlers

import (
	"net/http"

	"github.com/minhnguyen/pkg/config"
	"github.com/minhnguyen/pkg/models"
	"github.com/minhnguyen/pkg/render"
)

type Repository struct {
	AppConfig *config.AppConfig
}

var Repo *Repository

func NewRepository(a *config.AppConfig) *Repository {
	return &Repository{
		AppConfig: a,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.AppConfig.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) GeneralRoom(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "general.page.tmpl", &models.TemplateData{})
}

func (m *Repository) SuiteRoom(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "suite.page.tmpl", &models.TemplateData{})

}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "contact.page.tmpl", &models.TemplateData{})

}

func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "search-availability.page.tmpl", &models.TemplateData{})

}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, "make_reservation.page.tmpl", &models.TemplateData{})

}
