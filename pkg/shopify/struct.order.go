package shopify

import (
	"encoding/json"
	"time"
)

type Order struct {
	raw                                  []byte
	ID                                   string         `json:"id"`
	OrderNumber                          string         `json:"order_number"`
	Customer                             Customer       `json:"customer"`
	BillingAddress                       MailingAddress `json:"billingAddress"`
	BillingAddressMatchesShippingAddress bool           `json:"billingAddressMatchesShippingAddress"`
	ShippingAddress                      MailingAddress `json:"shippingAddress"`
	PaymentTerms                         PaymentTerms   `json:"paymentTerms"`
	Phone                                string         `json:"phone"`
	Email                                string         `json:"email"`
	CreatedAt                            time.Time      `json:"createdAt"`
	ProcessedAt                          time.Time      `json:"processedAt"`
	ClosedAt                             time.Time      `json:"closedAt"`
	SubtotalPriceSet                     PriceSet       `json:"subtotalPriceSet"`
	TotalTaxSet                          PriceSet       `json:"totalTaxSet"`
	TotalShippingPriceSet                PriceSet       `json:"totalShippingPriceSet"`
	TotalPriceSet                        PriceSet       `json:"totalPriceSet"`
	TotalReceivedSet                     PriceSet       `json:"totalReceivedSet"`
	LineItems                            []LineItem     `json:"lineItems"`
	DisplayFulfillmentStatus             string         `json:"displayFulfillmentStatus"`
	DisplayFinancialStatus               string         `json:"displayFinancialStatus"`
	Test                                 bool           `json:"test"`
	Closed                               bool           `json:"closed"`
}

type PresentmentMoney struct {
	Amount      float64 `json:"amount,string"`
	CurrentCode string  `json:"currencyCode"`
}

type PriceSet struct {
	PresentmentMoney PresentmentMoney `json:"presentmentMoney"`
}

type PaymentTerms struct {
	Name string `json:"paymentTermsName"`
	Type string `json:"paymentTermsType"`
}

type LineItem struct {
	ID       string  `json:"id"`
	Product  Product `json:"product"`
	Variant  Variant `json:"variant"`
	Quantity int     `json:"quantity"`
}

func (o Order) Raw() string {
	return string(o.raw)
} // ./Raw

func (o *Order) UnmarshalJSON(data []byte) error {
	type order struct {
		ID                                   string         `json:"id"`
		OrderNumber                          string         `json:"order_number"`
		Customer                             Customer       `json:"customer"`
		BillingAddress                       MailingAddress `json:"billingAddress"`
		BillingAddressMatchesShippingAddress bool           `json:"billingAddressMatchesShippingAddress"`
		ShippingAddress                      MailingAddress `json:"shippingAddress"`
		PaymentTerms                         PaymentTerms   `json:"paymentTerms"`
		Phone                                string         `json:"phone"`
		Email                                string         `json:"email"`
		CreatedAt                            time.Time      `json:"createdAt"`
		ProcessedAt                          time.Time      `json:"processedAt"`
		ClosedAt                             time.Time      `json:"closedAt"`
		SubtotalPriceSet                     PriceSet       `json:"subtotalPriceSet"`
		TotalTaxSet                          PriceSet       `json:"totalTaxSet"`
		TotalShippingPriceSet                PriceSet       `json:"totalShippingPriceSet"`
		TotalPriceSet                        PriceSet       `json:"totalPriceSet"`
		TotalReceivedSet                     PriceSet       `json:"totalReceivedSet"`
		LineItems                            struct {
			Nodes []LineItem `json:"nodes"`
		} `json:"lineItems"`
		DisplayFulfillmentStatus string `json:"displayFulfillmentStatus"`
		Test                     bool   `json:"test"`
		Closed                   bool   `json:"closed"`
	}
	var _o order
	err := json.Unmarshal(data, &_o)
	if err != nil {
		return err
	}
	*o = Order{
		raw:                                  data,
		ID:                                   _o.ID,
		OrderNumber:                          _o.OrderNumber,
		Customer:                             _o.Customer,
		BillingAddress:                       _o.BillingAddress,
		BillingAddressMatchesShippingAddress: _o.BillingAddressMatchesShippingAddress,
		ShippingAddress:                      _o.ShippingAddress,
		PaymentTerms:                         _o.PaymentTerms,
		Phone:                                _o.Phone,
		Email:                                _o.Email,
		CreatedAt:                            _o.CreatedAt,
		ProcessedAt:                          _o.ProcessedAt,
		ClosedAt:                             _o.ClosedAt,
		SubtotalPriceSet:                     _o.SubtotalPriceSet,
		TotalTaxSet:                          _o.TotalTaxSet,
		TotalShippingPriceSet:                _o.TotalShippingPriceSet,
		TotalPriceSet:                        _o.TotalPriceSet,
		TotalReceivedSet:                     _o.TotalReceivedSet,
		LineItems:                            _o.LineItems.Nodes,
		DisplayFulfillmentStatus:             _o.DisplayFulfillmentStatus,
		Test:                                 _o.Test,
		Closed:                               _o.Closed,
	}
	return nil
}
