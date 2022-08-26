package product

import (
	"context"
	"log"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx"
	"github.com/mittz/roleplay-webapp-assess/database"
)

const (
	IMAGEHASHES_DATA_FILENAME = "image_hashes.json"
)

type ImageHash struct {
	Name string
	Hash string
}

func GetNumOfProducts() int {
	dbPool := database.GetDatabaseConnection()

	var products []*ImageHash
	if err := pgxscan.Select(context.Background(), dbPool, &products, `SELECT name, hash from image_hashes`); err != nil {
		log.Println(err)
	}

	return len(products)
}

func GetImageHash(key string) string {
	dbPool := database.GetDatabaseConnection()

	product := ImageHash{}
	if err := dbPool.QueryRow(context.Background(), "select name, hash from image_hashes where name=$1", key).Scan(
		&product.Name,
		&product.Hash,
	); err != nil && err != pgx.ErrNoRows {
		log.Printf("QueryRow failed: %v\n", err)
	}

	return product.Hash
}
