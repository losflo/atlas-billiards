package shopify

type UserErrors struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}
