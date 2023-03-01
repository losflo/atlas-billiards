package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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

	fOut, err := os.OpenFile("not-shipped.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fOut.Close()

	w := csv.NewWriter(fOut)
	w.Write([]string{
		"Order Number",
		"SKU",
		"Quantity",
	})
	w.Flush()

	type response struct {
		Order struct {
			shopify.Order
			Refunds []struct {
				RefundLineItems struct {
					Nodes []struct {
						LineItem shopify.LineItem `json:"lineItem"`
					} `json:"nodes"`
				} `json:"refundLineItems"`
			} `json:"refunds"`
		} `json:"order"`
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
		// rows[0] = "5105750507771"
		oid := fmt.Sprintf("gid://shopify/Order/%s", rows[0])

		rq := graphql.NewRequest(fmt.Sprintf(`
			{
				order(id:"%s"){
					id
					order_number:name
					customer{
					  id
					  firstName
					  lastName
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
					refunds(first: 100){
						refundLineItems(first: 100){
						  nodes{
							lineItem{
							  sku
							  variantTitle
							  originalUnitPriceSet{
								presentmentMoney{
								  amount
								}
							  }
							  nonFulfillableQuantity
							}
						  }
						}
					}
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
		for _, rfs := range rs.Order.Refunds {
			for _, v := range rfs.RefundLineItems.Nodes {
				w.Write([]string{
					strings.ReplaceAll(rs.Order.OrderNumber, "#", ""),
					v.LineItem.Sku,
					strconv.Itoa(v.LineItem.NonFulfillableQuantity),
				})
				w.Flush()
			}
		}
	}
}
