package main

import (
	"context"
	"fmt"
	"imports/database"
	"imports/direxion"
	"imports/models"
)

func orchestrate(ctx context.Context, db database.DB, client direxion.Client) error {
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

	fmt.Print(allHoldings)
	return nil
}

func main() {
	db := database.NewDumbDatabase()
	client, err := direxion.NewDirexionClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = orchestrate(context.Background(), db, client)
	if err != nil {
		fmt.Println(err)
		return
	}
}
