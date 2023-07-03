package dbrepo

import (
	"database/sql"

	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/repository"
)

type pgDBRepo struct {
	App *config.AppConfig
	DB *sql.DB
}

func NewPostgresRepo(conn *sql.DB, app *config.AppConfig) repository.DatabaseRepo {
	return &pgDBRepo{
		app, 
		conn,
	}
}