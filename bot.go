package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/soulsplit/goex"
)

// TODO: decide on or get a pair to trade on
// TODO: define negative change in percent to set buy price
// TODO: define postive change in percent to set sell price
// TODO:
// TODO:

func autoTrade(pair goex.CurrencyPair, buyPercent float64, sellPercent float64, amount string) {
	price := getPrice(pair)

	// sample:
	// current price $12
	// want to buy at a 10% loss => $10.80
	// calculation: $12 - $12 * 10/100 = $12 - $1.20 = 10.80
	targetBuyPrice, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", price.Last-(price.Last*buyPercent/100.0)), 64)

	// sample:
	// price when bought: $10.80
	// want to sell at a 20% win => $10.80
	// calculation: $10.80 * (1 + 20/100) = $10.80 * 1,20 = 12,96
	targetSellPrice, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", targetBuyPrice*(1+sellPercent/100.0)), 64)

	orderBuy := createOrder(pair, amount, fmt.Sprintf("%f", targetBuyPrice), "LimitBuy")
	fmt.Printf("Buy order created: %s\n", orderBuy.OrderID2)

	// monitor the trade and wait for the order to be gone
	monitorTrade(orderBuy)
	fmt.Printf("Buy order done: %s\n", orderBuy.OrderID2)

	// show full order table
	var fiat = goex.Currency{Symbol: "EUR", Desc: ""}
	var orderMap = make(openOrdersMap)
	orderMap.getOpenOrders(fiat)
	printOrdersTable(*&orderMap)

	// if the buy order was fill, set the sell order
	orderSell := createOrder(pair, amount, fmt.Sprintf("%f", targetSellPrice), "LimitSell")
	fmt.Printf("Sell order created: %s\n", orderSell.OrderID2)
	monitorTrade(orderSell)
	fmt.Printf("Sell order done: %s\n", orderSell.OrderID2)
	fmt.Printf("Difference made: %f\n", targetSellPrice-targetBuyPrice)
}

func monitorTrade(order *goex.Order) {
	orderPresent := false
	var fiat = goex.Currency{Symbol: "EUR", Desc: ""}
	var orderMap = make(openOrdersMap)
	orderMap.getOpenOrders(fiat)
	for _, orders := range orderMap {
		for _, item := range orders {
			if item.orderID == order.OrderID2 {
				// once the order is filled, this loop can be left
				orderPresent = true
				break
			}
		}
		// order is still present
		if orderPresent {
			break
		}
	}
	// order is still present, so let's wait a bit and check again
	if orderPresent {
		fmt.Printf("Order not filled yet: %s\n", order.OrderID2)
		time.Sleep(time.Duration(60) * time.Second)
		monitorTrade(order)
	} else {
		fmt.Printf("Order filled yet: %s\n", order.OrderID2)
	}
}
