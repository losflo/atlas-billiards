package shopify

type Product struct {
	Title  string `json:"title"`
	Handle string `json:"handle"`
}

type Variant struct {
	ID                string  `json:"id"`
	DisplayName       string  `json:"displayName"`
	Title             string  `json:"title"`
	Sku               string  `json:"sku"`
	Price             string  `json:"price"`
	Weight            float64 `json:"weight"`
	InventoryQuantity int     `json:"inventoryQuantity"`
}
