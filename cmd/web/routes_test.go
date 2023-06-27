package main

import (
	"fmt"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/minhnguyen/internal/config"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		//do nothing
	default:
		t.Error(fmt.Sprintf("Failed. return type is not *chi.mux but is %T", v))
	}
}