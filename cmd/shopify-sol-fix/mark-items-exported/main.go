package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"atlasbilliards.com/pkg/shopify"
	"github.com/machinebox/graphql"
)

func main() {
	token := os.Getenv("ATLAS_BILLIARDS_SHOPIFY_ACCESS_TOKEN")
	endpoint := fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01/graphql.json", os.Getenv("ATLAS_BILLIARDS_SHOPIFY_SHOP"))
	client := graphql.NewClient(endpoint)

	f, err := os.OpenFile("../orders.csv", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	type response struct {
		Orders struct {
			Edges []struct {
				Order shopify.Order `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				StartCursor string `json:"startCursor"`
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"orders"`
	}

	r := csv.NewReader(f)
	for {
		rows, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		oid := rows[0]

		rq := graphql.NewRequest(fmt.Sprintf(`
			{
				order(id:%s){
					id
					order_number:name
					customer{
						id
						email
						firstName
						lastName
						defaultAddress{
							address1
							address2
							city
							state:provinceCode
							zip
							countryCodeV2
							region:provinceCode
							company
						}
						addresses{
							address1
							address2
							city
							state:provinceCode
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
						tax_exempt_id:metafield(namespace: "custom", key: "tax_exempt_id") {
							value
						}
						tags
						createdAt
					}
					billingAddress{
						firstName
						lastName
						phone
						address1
						address2
						city
						state:provinceCode
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
						state:provinceCode
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
					totalReceivedSet {
						presentmentMoney {
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
					displayFinancialStatus
					displayFulfillmentStatus
					closed
				}
			}
		`, oid))
		rq.Header.Add("X-Shopify-Access-Token", token)
		var rs response
		// var i GetRaw
		err = client.Run(context.Background(), rq, &rs)
		if err != nil {
			panic(err)
		}
	}
}
