package shopify

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"atlasbilliards.com/pkg/date"
	"github.com/machinebox/graphql"
)

type apiMeta struct {
	accessToken   string
	shop          string
	endpoint      string
	adminEndpoint string
	locationID    string
}

type Service struct {
	apiMeta
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
		apiMeta: apiMeta{
			accessToken:   conf.AccessToken,
			shop:          conf.Shop,
			endpoint:      fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01/graphql.json", conf.Shop),
			adminEndpoint: fmt.Sprintf("https://%s.myshopify.com/admin/api/2023-01", conf.Shop),
			locationID:    "gid://shopify/Location/71752646907",
		},
	}
} // ./NewService

func (s Service) writeMembersLine(c Customer, w *csv.Writer) error {
	tags := map[string]bool{}
	for _, v := range c.Tags {
		tags[v] = true
	}
	priceClass := "Retail"
	if _, ok := tags["wholesale"]; ok {
		priceClass = "Wholesale"
	}
	a := MailingAddress{}
	if c.DefaultAddress != nil {
		a = *c.DefaultAddress
	}
	company := ""
	if a.Company == "NULL" || a.Company == "Null" || a.Company == "null" {
		company = ""
	} else {
		company = a.Company
	}
	sepMem := strings.Split(c.ID, "/")
	memID := sepMem[len(sepMem)-1]
	err := w.Write([]string{
		memID,
		c.CustomerNumber.Value,
		c.Email,
		c.FirstName,
		c.LastName,
		company,
		a.Address1,
		ReplaceNull(a.Address2),
		a.City,
		a.State,
		a.Zip,
		a.Country,
		a.Region,
		FormatPhone(c.Phone),
		"",
		FormatPhone(c.Phone),
		"",
		priceClass,
		"Yes",
		date.ToSolomonDateFormat(c.CreatedAt),
		"",
		"",
	})
	if err != nil {
		return err
	}
	w.Flush()
	return nil
} // ./writeMembersLine

func (s Service) writeLineItems(orderNumber string, ll []LineItem, w *csv.Writer) error {
	for _, l := range ll {
		sepLI := strings.Split(l.ID, "/")
		LIID := sepLI[len(sepLI)-1]
		// id := sep[len(sep)-1]
		w.Write([]string{
			LIID,
			strings.Replace(orderNumber, "#", "", -1),
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
		w.Flush()
	}
	return nil
} // ./writeLineItems

func (s Service) updateCustomerMetafields(c Customer) error {
	client := graphql.NewClient(s.endpoint)
	rq := graphql.NewRequest(`
		mutation updateCustomerMetafields($input: CustomerInput!) {
			customerUpdate(input: $input) {
				customer {
					id
					metafields(first: 3) {
						edges {
							node {
								id
								namespace
								key
								value
							}
						}
					}
				}
				userErrors {
					message
					field
				}
			}
	  	}
	`)
	type metafields struct {
		ID  string `json:"id,omitempty"`
		Ns  string `json:"namespace"`
		Key string `json:"key"`
		Val string `json:"value"`
	}
	type input struct {
		Metafields []metafields `json:"metafields"`
		ID         string       `json:"id"`
	}
	in := input{
		Metafields: []metafields{
			{
				ID:  c.CustomerNumber.ID,
				Ns:  "custom",
				Key: "customer_number",
				Val: c.CustomerNumber.Value,
			},
			{
				ID:  c.TaxExemptID.ID,
				Ns:  "custom",
				Key: "tax_exempt_id",
				Val: c.TaxExemptID.Value,
			},
		},
		ID: c.ID,
	}
	rq.Var("input", in)
	rq.Header.Add("X-Shopify-Access-Token", s.accessToken)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	type response struct {
		CustomerUpdate *struct {
			Customer Customer `json:"customer"`
		} `json:"customerUpdate"`
		UserErrors struct {
			Message string `json:"message"`
			Field   string `json:"field"`
		} `json:"userErrors"`
	}
	var i GetRaw
	// var i response

	err := client.Run(ctx, rq, &i)
	if err != nil {
		return err
	}
	fmt.Println(string(i.Raw))
	return nil
} // ./updateCustomerMetafields

func (s Service) updateOrderTags(o Order, tags []string) error {
	return nil
} // ./updateOrderTags

func (s Service) OrderClosedAddArchiveTag() error {
	return nil
} // ./OrderClosedAddArchiveTag

func (s Service) SolomonMembersMapMetafields() error {
	type custInfo struct {
		CustomerNumber string
		TaxID          string
	}

	emailsMap := map[string]custInfo{}
	addressMap := map[string]custInfo{}

	fsol, err := os.OpenFile("solomon_members_clean.csv", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fsol.Close()
	rsol := csv.NewReader(fsol)
	for {
		/*
			CustomerNumber: 0
			LastName: 1
			Address1: 2
			Address2: 3
			City: 4
			State: 5
			Zip: 6
			Country: 7
			ResaleCertificateNumber: 8
			ResaleState: 9
			Email: 10
		*/
		rows, err := rsol.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		email := strings.ToLower(rows[10])
		if email == "test@cuestik.com" {
			continue
		}
		emailsMap[email] = custInfo{
			CustomerNumber: rows[0],
			TaxID:          rows[8],
		}

		addr := strings.ToLower(rows[2])
		addressMap[addr] = custInfo{
			CustomerNumber: rows[0],
			TaxID:          rows[8],
		}
	}

	client := graphql.NewClient(s.endpoint)
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
	ctx := context.Background()
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
								id
								value
							}
							tax_exempt_id:metafield(namespace: "custom", key: "tax_exempt_id") {
								id
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

		for _, e := range i.Customers.Edges {
			c := e.Customer
			tags := map[string]bool{}
			for _, v := range c.Tags {
				tags[v] = true
			}
			a := MailingAddress{}
			if c.DefaultAddress != nil {
				a = *c.DefaultAddress
			}

			custNumber := ""
			taxID := ""
			found := false

			if v, ok := emailsMap[strings.ToLower(c.Email)]; ok {
				found = true
				if c.CustomerNumber.Value != "" {
					custNumber = c.CustomerNumber.Value
				} else {
					custNumber = v.CustomerNumber
				}
				if c.TaxExemptID.Value != "" {
					taxID = c.CustomerNumber.Value
				} else {
					taxID = v.TaxID
				}
			} else {
				// email not found try address match
				// fmt.Printf("%s not found\n", strings.ToLower(c.Email))
				rgx := regexp.MustCompile(`\s\s+`)
				cleanAddr := a.Address1
				cleanAddr = rgx.ReplaceAllString(cleanAddr, " ")
				cleanAddr = strings.TrimSpace(strings.ToLower(cleanAddr))
				if _, ok := addressMap[cleanAddr]; ok {
					found = true
					if c.CustomerNumber.Value != "" {
						custNumber = c.CustomerNumber.Value
					} else {
						custNumber = v.CustomerNumber
					}
					if c.TaxExemptID.Value != "" {
						taxID = c.CustomerNumber.Value
					} else {
						taxID = v.TaxID
					}
				}
			}
			// not found
			// check if metafields already set
			// if set, replace custNumber and taxID
			if !found {
				if c.CustomerNumber.Value != "" {
					custNumber = c.CustomerNumber.Value

				}
				if c.TaxExemptID.Value != "" {
					taxID = c.TaxExemptID.Value
				}
			}
			if strings.ToUpper(custNumber) == "NULL" {
				custNumber = ""
			}
			if taxID != "" {
				cleantid := strings.ToLower(strings.ReplaceAll(taxID, " ", ""))
				if cleantid == "idonothaveone" {
					taxID = ""
				}
				if strings.ToUpper(taxID) == "NULL" {
					taxID = ""
				}
			}
			c.CustomerNumber.Value = custNumber
			c.TaxExemptID.Value = taxID
			if i.Customers.PageInfo.HasNextPage {
				after = fmt.Sprintf(" after: \"%s\"", i.Customers.PageInfo.EndCursor)
			}

			hasNextPage = i.Customers.PageInfo.HasNextPage
			err = s.updateCustomerMetafields(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
} // ./SolomonMembersMapMetafields

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
				c.Customer.ID,
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

func (s Service) GenSolonomFiles() error {
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

	// init members items file
	fMembers, err := os.OpenFile("MEMBERS.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fMembers.Close()

	wMembers := csv.NewWriter(fMembers)
	wMembers.Write([]string{
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
	wMembers.Flush()

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
				orders(first:1%s, query:"test:false AND fulfillment_status:fulfilled AND -financial_status:authorized AND tag_not:exported AND tag_not:archived"){
					edges{
						node{
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
					pageInfo{
						hasNextPage
						endCursor
					}
				}
			}
		`, after))
		rq.Header.Add("X-Shopify-Access-Token", s.accessToken)
		var rs response
		// var i GetRaw
		err := client.Run(ctx, rq, &rs)
		if err != nil {
			return err
		}
		for _, e := range rs.Orders.Edges {
			o := e.Order
			// if o.Closed {
			// 	continue
			// }
			if o.DisplayFulfillmentStatus != "FULFILLED" {
				continue
			}
			if o.Test {
				continue
			}
			c := o.Customer
			err = s.writeMembersLine(c, wMembers)
			if err != nil {
				return err
			}
			billA := o.BillingAddress
			shipA := o.ShippingAddress
			if o.BillingAddressMatchesShippingAddress {
				billA = shipA
			}
			sep := strings.Split(o.ID, "/")
			id := sep[len(sep)-1]

			sepMem := strings.Split(c.ID, "/")
			memID := sepMem[len(sepMem)-1]

			err := wOrders.Write([]string{
				id,
				c.CustomerNumber.Value,
				strings.Replace(o.OrderNumber, "#", "", -1),
				"WEB",
				memID, // MEMBER ID
				billA.FirstName,
				billA.LastName,
				billA.Company,
				billA.Address1,
				billA.Address2,
				billA.City,
				billA.State,
				billA.Country,
				billA.Zip,
				FormatPhone(billA.Phone),
				shipA.FirstName,
				shipA.LastName,
				shipA.Company,
				shipA.Address1,
				shipA.Address2,
				shipA.City,
				shipA.State,
				shipA.Country,
				shipA.Zip,
				FormatPhone(shipA.Phone),
				"NA", // SHIPPING CODE
				"CC", // TERMS
				c.Email,
				fmt.Sprintf("%.2f", o.SubtotalPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.SubtotalPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalTaxSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalShippingPriceSet.PresentmentMoney.Amount),
				fmt.Sprintf("%.2f", o.TotalPriceSet.PresentmentMoney.Amount),
				date.ToSolomonDateFormat(o.CreatedAt),
				date.ToSolomonDateFormat(o.CreatedAt), // PROCESS DATE TODO: TEMP
				date.ToSolomonDateFormat(o.ClosedAt),
				"",                                    // INVOICED_DATE
				date.ToSolomonDateFormat(o.CreatedAt), // SHIPPED DATE TODO: TEMP
				"",                                    // SMALL ORDER FEE
				"",                                    // LARGE ORDER DISCOUNT
			})
			if err != nil {
				return err
			}
			wOrders.Flush()

			err = s.writeLineItems(o.OrderNumber, o.LineItems, wCartItems)
			if err != nil {
				return err
			}
		}
		if rs.Orders.PageInfo.HasNextPage {
			after = fmt.Sprintf(" after: \"%s\"", rs.Orders.PageInfo.EndCursor)
		}
		hasNextPage = rs.Orders.PageInfo.HasNextPage
	}
	return nil
} // ./GenSolonomFiles

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

func (s Service) UploadInventory() error {
	f, err := os.OpenFile("ABS Inventory Quantities.txt", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fnis, err := os.OpenFile("not_in_shopify.csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fnis.Close()
	w := csv.NewWriter(fnis)

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	first := true
	for {
		/*
			InventoryID: 0
			Description: 1
			StockingUOM: 2
			PurchasingUOM: 3
			SellingUOM: 4
			StatusCode: 5
			Quantity: 6
		*/
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if first {
			w.Write(row)
			first = false
			continue
		}
		sku := strings.TrimSpace(row[0])
		ii, err := s.inventoryItemBySku(sku)
		if err != nil {
			fmt.Printf("sku %s: %s\n", sku, err.Error())
			w.Write(row)
			continue
		}
		if ii == nil {
			fmt.Printf("sku %s not in Shopify\n", sku)
			w.Write(row)
			continue
		}
		quantity, err := strconv.Atoi(row[6])
		if err != nil {
			return err
		}
		quantity = 1
		fmt.Println(quantity)
		err = ii.UpdateQuantity(quantity)
		if err != nil {
			return err
		}
		break
	}
	return nil
} // ./UploadInventory

func (s Service) inventoryItemBySku(sku string) (*InventoryItem, error) {
	client := graphql.NewClient(s.endpoint)
	client.Log = func(s string) { log.Println(s) }
	rq := graphql.NewRequest(fmt.Sprintf(`
		{
			inventoryItems(query: "sku:'%s'", first: 10) {
				edges{
					node{
						id
						sku
						inventoryLevel(locationId: "%s"){
							id
							available
						}
					}
				}
			}
		}
	`, sku, s.locationID))
	type response struct {
		InventoryItems struct {
			Edges []struct {
				Node InventoryItem `json:"node"`
			} `json:"edges"`
		} `json:"inventoryItems"`
	}
	rq.Header.Add("X-Shopify-Access-Token", s.accessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var res response
	err := client.Run(ctx, rq, &res)
	if err != nil {
		return nil, err
	}
	if len(res.InventoryItems.Edges) > 1 {
		fmt.Println("duplicate skus for ", sku)
	}
	if len(res.InventoryItems.Edges) == 0 {
		return nil, nil
	}
	ii := res.InventoryItems.Edges[0].Node
	ii.apiMeta = s.apiMeta
	return &ii, nil
} // ./InventoryItemBySku

func FormatPhone(s string) string {
	return strings.Replace(s, "+1", "", -1)
} // ./FormatPhone

func ReplaceNull(s string) string {
	return strings.Replace(s, "NULL", "", -1)
} // ./ReplaceNull
