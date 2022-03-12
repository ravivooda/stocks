package main

import (
	"context"
	"imports/database"
	"imports/direxion"
	"imports/models"
)

func orchestrate(ctx context.Context, db database.DB, client direxion.DirexionClient) error {
	seeds, err := db.ListSeeds(ctx)
	if err != nil {
		return err
	}

	var allHoldings []models.Holding
	for _, seed := range seeds {
		holdings, err := client.GetHoldings(ctx, seed)
		if err != nil {
			return err
		}
		allHoldings = append(allHoldings, holdings...)
	}

	print(allHoldings)
	return nil
}

func main() {
	db := database.NewDumbDatabase()
	client, err := direxion.NewDirexionClient()
	if err != nil {
		return
	}

	err = orchestrate(context.Background(), db, client)
	if err != nil {
		return
	}
}
