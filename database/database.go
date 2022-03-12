package database

import (
	"context"
	"database/sql"
	"imports/models"
)

type DBConfig struct {
	DriverName string
	Location   string
}

type DB interface {
	ListSeeds(context context.Context) ([]models.Seed, error)
}

type importsDB struct {
	db *sql.DB
}

func (i importsDB) ListSeeds(ctx context.Context) ([]models.Seed, error) {
	rows, err := i.db.QueryContext(ctx, "SELECT url, ticker, header FROM seeds")
	if err != nil {
		return nil, err
	}
	var rets []models.Seed
	for rows.Next() {
		seed := models.Seed{}
		err := rows.Scan(&seed.URL, &seed.Ticker, &seed.Header.SkippableLines, &seed.Header.ExpectedColumns)
		if err != nil {
			return nil, err
		}
		rets = append(rets, seed)
	}
	return rets, nil
}

func NewDB(config DBConfig) (DB, error) {
	db, err := sql.Open(config.DriverName, config.Location)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &importsDB{
		db: db,
	}, nil
}
