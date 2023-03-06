package main

import (
	"os"

	"atlasbilliards.com/pkg/shopify"
)

func main() {
	var err error
	conf := shopify.Config{
		AccessToken: os.Getenv("ATLAS_BILLIARDS_SHOPIFY_ACCESS_TOKEN"),
		Shop:        os.Getenv("ATLAS_BILLIARDS_SHOPIFY_SHOP"),
	}
	s := shopify.NewService(conf)
	query := "test:false AND fulfillment_status:fulfilled AND -financial_status:authorized AND tag_not:exported AND tag_not:archived AND tag:printed AND created_at:2023-02-28"
	err = s.GenSolonomFiles(query)
	if err != nil {
		panic(err)
	}
}
