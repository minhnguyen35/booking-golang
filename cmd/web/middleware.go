package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func NoSurf(next http.Handler) http.Handler {
	csrf := nosurf.New(next)

	csrf.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrf
}

func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}