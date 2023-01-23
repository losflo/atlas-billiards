package shopify

type GetRaw struct {
	Raw []byte
}

func (g *GetRaw) UnmarshalJSON(data []byte) error {
	g.Raw = data
	return nil
}
