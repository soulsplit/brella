package main

import (
	"fmt"
	"sort"

	"github.com/soulsplit/goex"
)

// retieve the list of assets as a list of currency pairs offered by the exchange
func getAllAssets() []goex.CurrencyPair {
	curr := goex.Currency{Symbol: "", Desc: ""}
	p := getPair("all", curr)
	res, err := apiGoex.GetAssets(p)
	if err != nil {
		fmt.Println(err)
	}
	return res.Assets
}

// retieve the list of assets as a list of strings offered by the exchange
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

// do the reverse lookup from a string to a currency pair
func assetStringToCurrencyObject(currencypair string) goex.CurrencyPair {
	for _, v := range krakenAssets {
		if currencypair == v.String() {
			return v
		}
	}
	return goex.CurrencyPair{}
}
