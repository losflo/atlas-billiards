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
	CurrentSubtotalPriceSet              PriceSet       `json:"currentSubtotalPriceSet"`
	CurrentTotalTaxSet                   PriceSet       `json:"currentTotalTaxSet"`
	TotalShippingPriceSet                PriceSet       `json:"totalShippingPriceSet"`
	CurrentTotalPriceSet                 PriceSet       `json:"currentTotalPriceSet"`
	TotalReceivedSet                     PriceSet       `json:"totalReceivedSet"`
	NetPaymentSet                        PriceSet       `json:"netPaymentSet"`
	LineItems                            []LineItem     `json:"lineItems"`
	Fulfillments                         []Fulfillment  `json:"fulfillments"`
	DisplayFulfillmentStatus             string         `json:"displayFulfillmentStatus"`
	DisplayFinancialStatus               string         `json:"displayFinancialStatus"`
	Tags                                 []string       `json:"tags"`
	Test                                 bool           `json:"test"`
	Closed                               bool           `json:"closed"`
}

type Fulfillment struct {
	FulfillmentLineItems struct {
		Nodes []FulfillmentLineItem `json:"nodes"`
	} `json:"fulfillmentLineItems"`
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
	ID                     string   `json:"id"`
	Product                Product  `json:"product"`
	Variant                Variant  `json:"variant"`
	Quantity               int      `json:"quantity"`
	CurrentQuantity        int      `json:"currentQuantity"`
	TotalQuantity          int      `json:"totalQuantity"`
	NonFulfillableQuantity int      `json:"nonFulfillableQuantity"`
	Sku                    string   `json:"sku"`
	VariantTitle           string   `json:"variantTitle"`
	OriginalUnitPriceSet   PriceSet `json:"originalUnitPriceSet"`
	DiscountedUnitPriceSet PriceSet `json:"discountedUnitPriceSet"`
}

type Refund struct {
	RefundLineItems []LineItem `json:"refundLineItems"`
}

type FulfillmentLineItem struct {
	LineItem           LineItem `json:"lineItem"`
	DiscountedTotalSet PriceSet `json:"discountedTotalSet"`
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
		CurrentSubtotalPriceSet              PriceSet       `json:"currentSubtotalPriceSet"`
		CurrentTotalTaxSet                   PriceSet       `json:"currentTotalTaxSet"`
		TotalShippingPriceSet                PriceSet       `json:"totalShippingPriceSet"`
		CurrentTotalPriceSet                 PriceSet       `json:"currentTotalPriceSet"`
		TotalReceivedSet                     PriceSet       `json:"totalReceivedSet"`
		NetPaymentSet                        PriceSet       `json:"netPaymentSet"`
		LineItems                            struct {
			Nodes []LineItem `json:"nodes"`
		} `json:"lineItems"`
		Fulfillments             []Fulfillment `json:"fulfillments"`
		DisplayFulfillmentStatus string        `json:"displayFulfillmentStatus"`
		Tags                     []string      `json:"tags"`
		Test                     bool          `json:"test"`
		Closed                   bool          `json:"closed"`
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
		CurrentSubtotalPriceSet:              _o.CurrentSubtotalPriceSet,
		CurrentTotalTaxSet:                   _o.CurrentTotalTaxSet,
		TotalShippingPriceSet:                _o.TotalShippingPriceSet,
		CurrentTotalPriceSet:                 _o.CurrentTotalPriceSet,
		TotalReceivedSet:                     _o.TotalReceivedSet,
		NetPaymentSet:                        _o.NetPaymentSet,
		LineItems:                            _o.LineItems.Nodes,
		Fulfillments:                         _o.Fulfillments,
		DisplayFulfillmentStatus:             _o.DisplayFulfillmentStatus,
		Tags:                                 _o.Tags,
		Test:                                 _o.Test,
		Closed:                               _o.Closed,
	}
	return nil
}
