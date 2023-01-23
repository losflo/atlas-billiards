package shopify

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"atlasbilliards.com/pkg/date"
	"github.com/machinebox/graphql"
)

type Service struct {
	accessToken string
	shop        string
	endpoint    string
}

type Config struct {
	Shop        string
	AccessToken string
}

func NewService(conf Config) *Service {
	if conf.Shop == "" || conf.AccessToken == "" {
		panic("Shop and AccessToken required")
	}
	return &Service{
		accessToken: conf.AccessToken,
		shop:        conf.Shop,
		endpoint:    fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01/graphql.json", conf.Shop),
	}
} // ./NewService

func (s Service) SolomonMembersExport() error {
	client := graphql.NewClient(s.endpoint)

	f, err := os.OpenFile("MEMBERS.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Write([]string{
		"MEMBER_ID",
		"CustId",
		"EMAIL",
		"FIRST_NAME",
		"LAST_NAME",
		"COMPANY_NAME",
		"ADDRESS1",
		"ADDRESS2",
		"CITY",
		"STATE_CODE",
		"ZIP",
		"COUNTRY_CODE",
		"REGION",
		"PHONE",
		"FAX",
		"CELL",
		"Terms",
		"PRICE_CLASS",
		"APPROVAL_PENDING",
		"DATE_CREATED",
		"LAST_UPDATED",
		"NOTES",
	})
	w.Flush()

	type response struct {
		Customers struct {
			Edges []struct {
				Customer Customer `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				StartCursor string `json:"startCursor"`
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"customers"`
	}

	var i response
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	hasNextPage := true
	after := ""
	for hasNextPage {
		rq := graphql.NewRequest(fmt.Sprintf(`
			{
				customers(first:150%s) {
					edges {
						node {
							id
							email
							firstName
							lastName
							defaultAddress{
								address1
								address2
								city
								state:province
								zip
								countryCodeV2
								region:provinceCode
								company
							}
							addresses{
								address1
								address2
								city
								state:province
								zip
								countryCodeV2
								region:provinceCode
								company
							}
							phone
							taxExempt
							taxExemptions
							customer_number:metafield(namespace: "custom", key:"customer_number") {
								value
							}
							tax_exempt_id:metafield(namespace: "custom", key: "tax_exept_id") {
								value
							}
							tags
							createdAt
						}
					}
					pageInfo {
						startCursor
						endCursor
						hasNextPage
					}
				}
			}
		`, after))
		rq.Header.Add("X-Shopify-Access-Token", s.accessToken)

		err = client.Run(ctx, rq, &i)
		if err != nil {
			return err
		}

		for _, c := range i.Customers.Edges {
			tags := map[string]bool{}
			for _, v := range c.Customer.Tags {
				tags[v] = true
			}
			priceClass := "Retail"
			if _, ok := tags["wholesale"]; ok {
				priceClass = "Wholesale"
			}
			a := MailingAddress{}
			if c.Customer.DefaultAddress != nil {
				a = *c.Customer.DefaultAddress
			}
			company := ""
			if a.Company == "NULL" || a.Company == "Null" || a.Company == "null" {
				company = ""
			} else {
				company = a.Company
			}
			err := w.Write([]string{
				"",
				c.Customer.CustomerNumber.Value,
				c.Customer.Email,
				c.Customer.FirstName,
				c.Customer.LastName,
				company,
				a.Address1,
				ReplaceNull(a.Address2),
				a.City,
				a.State,
				a.Zip,
				a.Country,
				a.Region,
				FormatPhone(c.Customer.Phone),
				"",
				FormatPhone(c.Customer.Phone),
				"",
				priceClass,
				"Yes",
				date.ToSolomonDateFormat(c.Customer.CreatedAt),
				"",
				"",
			})
			if err != nil {
				return err
			}
			w.Flush()
			if i.Customers.PageInfo.HasNextPage {
				after = fmt.Sprintf(" after: \"%s\"", i.Customers.PageInfo.EndCursor)
			}
			hasNextPage = i.Customers.PageInfo.HasNextPage
		}
	}
	return nil
} // ./SolomonMembersExport

func (s Service) SolomonStoreOrdersAndCartItemsExport() error {
	client := graphql.NewClient(s.endpoint)

	// init store orders file
	fOrders, err := os.OpenFile("STORE_ORDERS.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fOrders.Close()

	wOrders := csv.NewWriter(fOrders)
	wOrders.Write([]string{
		"ORDER_ID",
		"CustId",
		"ORDER_NR",
		"ADMIN_CODE",
		"MEMBER_ID",
		"BILLING_FIRST_NAME",
		"BILLING_LAST_NAME",
		"BILLING_COMPANY",
		"BILLING_ADDRESS1",
		"BILLING_ADDRESS2",
		"BILLING_CITY",
		"BILLING_STATE",
		"BILLING_COUNTRY",
		"BILLING_ZIP",
		"BILLING_PHONE",
		"SHIPPING_FIRST_NAME",
		"SHIPPING_LAST_NAME",
		"SHIPPING_COMPANY",
		"SHIPPING_ADDRESS1",
		"SHIPPING_ADDRESS2",
		"SHIPPING_CITY",
		"SHIPPING_STATE",
		"SHIPPING_COUNTRY",
		"SHIPPING_ZIP",
		"SHIPPING_PHONE",
		"SHIPPING_CODE",
		"Terms",
		"EMAIL",
		"BASE_SUBTOTAL",
		"SUBTOTAL",
		"TAX_AMOUNT",
		"SHIPPING_AMOUNT",
		"TOTAL",
		"CREATE_DATE",
		"PROCESS_DATE",
		"SETTLE_DATE",
		"INVOICED_DATE",
		"SHIPPED_DATE",
		"SMALL_ORDER_FEE",
		"LARGE_ORDER_DISCOUNT",
	})
	wOrders.Flush()

	// init store cart items file
	fCartItems, err := os.OpenFile("STORE_CART_ITEMS.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fCartItems.Close()

	wCartItems := csv.NewWriter(fCartItems)
	wCartItems.Write([]string{
		"CART_ITEM_ID",
		"ORDER_NR",
		"ITEM_VARIANT_ID",
		"ITEM_PRICE",
		"SALE_PRICE",
		"IS_ON_SALE",
		"ITEM_NUMBER",
		"UNIT_OF_MEASURE",
		"ITEM_QUANTITY",
		"ITEM_NAME",
		"WEIGHT",
		"PRICE",
		"EXTRA_PRICE",
		"OPTION_ID",
		"OPTION_ITEM_NUMBER_MODIFIER",
	})
	wCartItems.Flush()

	type response struct {
		Orders struct {
			Edges []struct {
				Order Order `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				StartCursor string `json:"startCursor"`
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"orders"`
	}

	hasNextPage := true
	after := ""

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	for hasNextPage {
		rq := graphql.NewRequest(fmt.Sprintf(`
			{
				orders(first:1%s){
					edges{
						node{
							id
							order_number:name
							customer{
								id
								email
								customer_number:metafield(namespace: "custom", key:"customer_number") {
									value
								}
								tax_exempt_id:metafield(namespace: "custom", key: "tax_exept_id") {
									value
								}
							}
							billingAddress{
								firstName
								lastName
								phone
								address1
								address2
								city
								state:province
								zip
								countryCodeV2
								region:provinceCode
								company
							}
							billingAddressMatchesShippingAddress
							shippingAddress{
								firstName
								lastName
								phone
								address1
								address2
								city
								state:province
								zip
								countryCodeV2
								region:provinceCode
								company
							}
							paymentTerms{
								paymentTermsName
								paymentTermsType
							}
							phone
							email
							createdAt
							processedAt
							closedAt
							subtotalPriceSet{
								presentmentMoney{
									amount
									currencyCode
								}
							}
							totalTaxSet{
								presentmentMoney{
									amount
									currencyCode
								}
							}
							totalShippingPriceSet{
								presentmentMoney{
									amount
									currencyCode
								}
							}
							totalPriceSet{
								presentmentMoney{
									amount
									currencyCode
								}
							}
							lineItems(first: 250) {
								nodes{
									id
									product{
										title
										handle
									}
									variant{
										title
										sku
										price
										weight
									}
									quantity
								}
							}
						}
					}
					pageInfo{
						hasNextPage
						endCursor
					}
				}
			}
		`, after))
		rq.Header.Add("X-Shopify-Access-Token", s.accessToken)
		var rs response
		err := client.Run(ctx, rq, &rs)
		if err != nil {
			return err
		}
		for _, e := range rs.Orders.Edges {
			o := e.Order
			c := o.Customer
			billA := o.BillingAddress
			shipA := o.ShippingAddress
			if o.BillingAddressMatchesShippingAddress {
				billA = shipA
			}
			sep := strings.Split(o.ID, "/")
			id := sep[len(sep)-1]
			err := wOrders.Write([]string{
				id,
				c.CustomerNumber.Value,
				strings.Replace(o.OrderNumber, "#", "", -1),
				"WEB",
				"", // TODO: MEMBER ID
				billA.FirstName,
				billA.FirstName,
				billA.Company,
				billA.Address1,
				billA.Address2,
				billA.City,
				billA.State,
				billA.Country,
				billA.Zip,
				FormatPhone(billA.Phone),
				shipA.FirstName,
				shipA.FirstName,
				shipA.Company,
				shipA.Address1,
				shipA.Address2,
				shipA.City,
				shipA.State,
				shipA.Country,
				shipA.Zip,
				FormatPhone(shipA.Phone),
				c.Email,
				fmt.Sprintf("%.2f", o.SubtotalPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.SubtotalPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalTaxSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalShippingPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalPriceSet.PresentmentMoney.Amount),
				date.ToSolomonDateFormat(o.CreatedAt),
				"", // TODO: PROCESS DATE
				date.ToSolomonDateFormat(o.ClosedAt),
				"", // TODO: SHIPPED DATE
				"", // TODO: SMALL ORDER FEE
				"", // LARGE ORDER DISCOUNT
			})
			if err != nil {
				return err
			}
			wOrders.Flush()
			for _, l := range o.LineItems {
				wCartItems.Write([]string{
					"",
					strings.Replace(o.OrderNumber, "#", "", -1),
					"", // TODO: ITEM VARIANT ID
					l.Variant.Price,
					"0.00",
					"False",
					l.Variant.Sku,
					"Each",
					fmt.Sprintf("%d", l.Quantity),
					l.Variant.Title,
					fmt.Sprintf("%.4f", l.Variant.Weight),
					l.Variant.Price,
					"", // EXTRA PRICE
					"", // OPTION ID
					"", // OPTION ID NUMBER MODIFIER
				})
				wCartItems.Flush()
			}
		}
		if rs.Orders.PageInfo.HasNextPage {
			after = fmt.Sprintf(" after: \"%s\"", rs.Orders.PageInfo.EndCursor)
		}
		hasNextPage = rs.Orders.PageInfo.HasNextPage
	}
	return nil
} // ./SolomongStoreOrdersAndCartItemsExport

func (s Service) SolomonInventoryExport() error {
	client := graphql.NewClient(s.endpoint)
	hasNextPage := true
	after := ""
	ii := []InventoryItem{}
	for hasNextPage {
		rq := graphql.NewRequest(fmt.Sprintf(`
			{
				inventoryItems(first: 250%s) {
					edges {
						cursor
						node {
							id
							sku
							unitCost {
								amount
								currencyCode
							}
							duplicateSkuCount
							variant {
								id
								displayName
								sku
								barcode
								inventoryQuantity
								price
							}
						}
					}
					pageInfo {
						startCursor
						endCursor
						hasNextPage
					}
				}
			}
		`, after))
		rq.Header.Add("X-Shopify-Access-Token", s.accessToken)
		type response struct {
			InventoryItems struct {
				Edges []struct {
					Cursor string
					Node   InventoryItem `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					StartCursor string `json:"startCursor"`
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"inventoryItems"`
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var rs response
		err := client.Run(ctx, rq, &rs)
		if err != nil {
			return err
		}
		for _, i := range rs.InventoryItems.Edges {
			ii = append(ii, i.Node)
		}
		if rs.InventoryItems.PageInfo.HasNextPage {
			after = fmt.Sprintf(" after: \"%s\"", rs.InventoryItems.PageInfo.EndCursor)
		}
		hasNextPage = rs.InventoryItems.PageInfo.HasNextPage
	}

	f, err := os.OpenFile("ABS Inventory Quantities.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = '\t'
	w.Write([]string{
		"InventoryID",
		"Description",
		"StockingUOM",
		"PurchasingUOM",
		"SellingUOM",
		"StatusCode",
		"Quantity",
	})
	w.Flush()
	for _, i := range ii {
		w.Write([]string{
			i.Variant.Sku,
			i.Variant.DisplayName,
			"EA",
			"EA",
			"EA",
			"AC",
			fmt.Sprintf("%d", i.Variant.InventoryQuantity),
		})
		w.Flush()
	}
	return nil
} // ./SolomonInventoryExport

func FormatPhone(s string) string {
	return strings.Replace(s, "+1", "", -1)
}

func ReplaceNull(s string) string {
	return strings.Replace(s, "NULL", "", -1)
} // ./ReplaceNull
