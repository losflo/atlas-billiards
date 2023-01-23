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
	// err = s.SolomonMembersExport()
	// if err != nil {
	// 	panic(err)
	// }
	// err = s.SolomonStoreOrdersAndCartItemsExport()
	// if err != nil {
	// 	panic(err)
	// }
	err = s.SolomonInventoryExport()
	if err != nil {
		panic(err)
	}
}
