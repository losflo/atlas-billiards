package shopify

type InventoryItem struct {
	ID                string   `json:"id"`
	Sku               string   `json:"sku"`
	UnitCost          UnitCost `json:"unitCost"`
	DuplicateSkuCount int      `json:"duplicateSkuCount"`
	Variant           Variant  `json:"variant"`
}

type UnitCost struct {
	Amount       interface{} `json:"amount"`
	CurrencyCode string      `json:"currencyCode"`
}
