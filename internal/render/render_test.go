package render

import (
	"net/http"
	"testing"

	"github.com/minhnguyen/internal/models"
)

func TestAddDefaultdata(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "123")

	result := AddDefaultData(&td, r)

	if result.Flash != "123" {
		t.Error("Failed AddDefaultData")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplate = "../../templates"
	tc, err := CreateTemplateCache()

	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww myWriter

	err = Template(&ww, r, "home.page.tmpl", &models.TemplateData{})

	if err != nil {
		t.Error("Cannot writing template to browser")
	}

	err = Template(&ww, r, "non-exist.page.tmpl", &models.TemplateData{})

	if err == nil {
		t.Error("Writing non-exist template to browser")
	}

}

func TestNewTemplate(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplate = "../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

type myWriter struct{}

func (w *myWriter) Header() http.Header {
	var h http.Header
	return h
}

func (w *myWriter) WriteHeader(n int) {

}

func (w *myWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}
