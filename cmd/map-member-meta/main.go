package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"
	"regexp"
	"strings"

	"atlasbilliards.com/pkg/shopify"
)

var clean bool

func init() {
	flag.BoolVar(&clean, "clean", false, "-clean")
	flag.Parse()
}

func main() {
	var err error
	// clean csv file
	if clean {
		fsol, err := os.OpenFile("solomon_members.csv", os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer fsol.Close()
		rsol := csv.NewReader(fsol)

		fclean, err := os.OpenFile("solomon_members_clean.csv", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer fclean.Close()
		wclean := csv.NewWriter(fclean)

		for {
			rows, err := rsol.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			rowsClean := []string{}
			for i, v := range rows {
				if i == 0 {
					rowsClean = append(rowsClean, v)
					continue
				}
				rgx := regexp.MustCompile(`\s\s+`)
				v2 := rgx.ReplaceAllString(v, " ")
				rowsClean = append(rowsClean, strings.ToLower(strings.TrimSpace(v2)))
			}
			wclean.Write(rowsClean)
			wclean.Flush()
		}
	}

	conf := shopify.Config{
		AccessToken: os.Getenv("ATLAS_BILLIARDS_SHOPIFY_ACCESS_TOKEN"),
		Shop:        os.Getenv("ATLAS_BILLIARDS_SHOPIFY_SHOP"),
	}
	s := shopify.NewService(conf)
	err = s.SolomonMembersMapMetafields()
	if err != nil {
		panic(err)
	}
}
