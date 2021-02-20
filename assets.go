package main

import (
	"fmt"
	"sort"

	"github.com/soulsplit/goex"
)

func getAllAssets() []goex.CurrencyPair {
	curr := goex.Currency{Symbol: "", Desc: ""}
	p := getPair("all", curr)
	res, err := apiGoex.GetAssets(p)
	if err != nil {
		fmt.Println(err)
	}
	return res.Assets
}

func getAllAssetsString() []string {
	var allAssets []string
	for _, v := range krakenAssets {
		if v.String() != "UNKNOWN_UNKNOWN" {
			allAssets = append(allAssets, v.String())
		}
	}
	sort.Strings(allAssets)
	return allAssets
}

func assetStringToCurrencyObject(currencypair string) goex.CurrencyPair {
	for _, v := range krakenAssets {
		if currencypair == v.String() {
			return v
		}
	}
	return goex.CurrencyPair{}
}
