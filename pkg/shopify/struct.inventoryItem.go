package shopify

import (
	"context"
	"time"

	"github.com/machinebox/graphql"
)

type InventoryItem struct {
	apiMeta
	ID                string          `json:"id"`
	Sku               string          `json:"sku"`
	UnitCost          UnitCost        `json:"unitCost"`
	DuplicateSkuCount int             `json:"duplicateSkuCount"`
	Variant           Variant         `json:"variant"`
	InventoryLevel    *InventoryLevel `json:"inventoryLevel"`
}

type InventoryLevel struct {
	ID        string `json:"id"`
	Available int    `json:"available"`
}

type UnitCost struct {
	Amount       interface{} `json:"amount"`
	CurrencyCode string      `json:"currencyCode"`
}

func (ii InventoryItem) UpdateQuantity(amountDelta int) error {
	rq := graphql.NewRequest(`
		mutation inventoryAdjustQuantity($input: InventoryAdjustQuantityInput!) {
			inventoryAdjustQuantity(input: $input) {
				inventoryLevel {
					id
					sku
					inventoryLevel(locationId: "%s"){
						id
						available
					}
				}
				userErrors {
					field
					message
				}
			}
		}
	`)
	input := map[string]interface{}{
		"inventoryLevelId": ii.ID,
		"availableDelta":   amountDelta,
	}
	rq.Var("input", input)
	client := graphql.NewClient(ii.endpoint)
	type response struct {
		InventoryAdjustQuantity struct {
			InventoryLevel InventoryLevel `json:"inventoryLevel"`
		} `json:"inventoryAdjustQuantity"`
	}
	var rs response
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Run(ctx, rq, &rs)
	if err != nil {
		return err
	}
	return nil
} // ./ UpdateQuantity
