package shopify

import (
	"encoding/json"
	"time"
)

type Customer struct {
	ID             string           `json:"id"`
	Email          string           `json:"email"`
	FirstName      string           `json:"firstName"`
	LastName       string           `json:"lastName"`
	DefaultAddress *MailingAddress  `json:"defaultAddress"`
	Addresses      []MailingAddress `json:"addresses"`
	Phone          string           `json:"phone"`
	TaxExempt      bool             `json:"taxExempt"`
	TaxExemptions  []string         `json:"taxExemptions"`
	CustomerNumber Metafield        `json:"customer_number"`
	TaxExemptID    Metafield        `json:"tax_exempt_id"`
	Tags           []string         `json:"tags"`
	CreatedAt      time.Time        `json:"createdAt"`
}

type Metafield struct {
	Value string `json:"value"`
}

func (m *Metafield) UnmarshalJSON(data []byte) error {
	if data == nil || len(data) <= 0 {
		*m = Metafield{}
		return nil
	}
	type avoidStackOverflow struct {
		Value string `json:"value"`
	}
	var a avoidStackOverflow
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	*m = Metafield(a)
	return nil
} // ./UnmarshalJSON

type MailingAddress struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"countryCodeV2"`
	Region    string `json:"province"`
	Company   string `json:"company"`
}
