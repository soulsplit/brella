package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/soulsplit/goex"
)

// map currency to a list of open orders
type openOrdersMap map[string][]openOrders

type openOrders struct {
	currency  string
	value     float64
	orderID   string
	amount    float64
	orderType goex.OrderType
}

func (odMap openOrdersMap) getOpenOrders(fiat goex.Currency) {
	// the currencyPair passed here is not evaluated for kraken. This call will get all open orders.
	openOrders, _ := apiGoex.GetUnfinishOrders(getPair("BTC", fiat))
	for _, order := range openOrders {
		odMap.storeOrders(order)
	}
}

func (odMap *openOrdersMap) toRows() [][]string {
	var formatted [][]string
	for _, orders := range *odMap {
		for _, item := range orders {
			formatted = append(formatted, []string{
				item.currency,
				strconv.FormatFloat(item.amount, 'f', 4, 64),
				strconv.FormatFloat(item.value, 'f', 2, 64),
				strconv.FormatFloat(item.value*item.amount, 'f', 2, 64),
				item.orderID})
		}
	}
	sort.Sort(matrix(formatted))
	return formatted
}

// add data to the map that holds all values from the current session
func (odMap openOrdersMap) storeOrders(order goex.Order) {
	currName := order.Currency.String()
	_, ok := odMap[currName]
	var orderSummary openOrders
	orderSummary.amount = order.Amount
	orderSummary.orderID = order.OrderID2
	orderSummary.currency = order.Currency.String()
	orderSummary.value = order.Price

	if !ok {
		odMap[currName] = []openOrders{orderSummary}
	} else {
		odMap[currName] = append(odMap[currName], orderSummary)
	}
}

func createOrder(pair goex.CurrencyPair, amount string, price string, orderType string) {
	if orderType == "LimitBuy" {
		_, err := apiGoex.LimitBuy(amount, price, pair)
		if err != nil {
			fmt.Println(err)
		}
	}
	var orderMap = make(openOrdersMap)
	orderMap.getOpenOrders(pair.CurrencyB)
	printOrdersTable(*&orderMap)
}

func deletelOrder(oderid string) {

	deleted, err := apiGoex.CancelOrder(oderid, goex.BTC_JPY) // currencypair is not needed by kraken api
	if err != nil {
		fmt.Println(err)
	}
	if deleted {
		fmt.Printf("Cancellation successful.")
	} else {
		fmt.Printf("Cancellation failed.")
	}
}
