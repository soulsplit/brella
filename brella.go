package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/soulsplit/goex"
	"github.com/soulsplit/goex/builder"
)

var apiGoex goex.API = getAPIHandle()

type matrix [][]string

type credentials struct {
	APIKey        string ""
	APISecretkey  string ""
	APIPassphrase string ""
}
type history struct {
	currName string ""
	value    []float64
}

type values map[string]float64

// map currency to a list of values
type valuesMap map[string][]float64

// map currency to a list of open orders
type openOrdersMap map[string][]openOrders

type openOrders struct {
	currency  string
	value     float64
	orderID   string
	amount    float64
	orderType goex.OrderType
}

type dataItem struct {
	currName         string
	price            float64
	value            float64
	amount           float64
	changeSinceStart string
	changeSinceLast  string
}

func (exData *exchangeData) toRows() [][]string {
	var formatted [][]string
	for _, item := range exData.items {
		formatted = append(formatted, []string{
			item.currName,
			strconv.FormatFloat(item.amount, 'f', 4, 64),
			strconv.FormatFloat(item.price, 'f', 2, 64),
			strconv.FormatFloat(item.value, 'f', 2, 64),
			item.changeSinceStart,
			item.changeSinceLast})
	}
	sort.Sort(matrix(formatted))
	return formatted
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

type exchangeData struct {
	items    []dataItem
	exchange string
	sum      float64
	fiat     goex.Currency
}

// Len() supports the sorting algorithm by providing the length of the array
func (s matrix) Len() int {
	return len(s)
}

// Swap() supports the sorting algorithm by providing the way how to change the order of elements in the array
func (s matrix) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less() supports the sorting algorithm by providing a way to identify which element is less/smaller than another in the array
func (s matrix) Less(i, j int) bool {
	return s[i][0] < s[j][0]
}

// Version is set
const Version string = "0.0.2"

// Print current version
func getVersion() {
	fmt.Println(Version)
}

// Cleans up the names that come from exchange APIs to make them conform, for example with ticker requests
func estimateName(name string) string {

	var newName string = name

	if strings.HasSuffix(name, ".S") {
		// stacked
		newName = strings.Trim(name, ".S")
	}
	if name == "EC" {
		// ZCash workaround for wrong name
		newName = "ZEC"
	}

	return newName
}

func (odMap openOrdersMap) getOpenOrders(fiat goex.Currency) {
	// the currencyPair passed here is not evaluated for kraken. This call will get all open orders.
	openOrders, _ := apiGoex.GetUnfinishOrders(getPair("BTC", fiat))
	for _, order := range openOrders {
		odMap.storeOrders(order)
	}
}

// extractHoldings() will pick out various parts of the user's balance data write a human-readable string
func (exData *exchangeData) extractHoldings(acc goex.Account, fiat goex.Currency, vMap valuesMap, ts string) {
	var value float64
	for _, subacc := range acc.SubAccounts {
		// skip small amounts, skip euro amount, skip redundant XBT as BTC is already reported for Bitcoin
		if subacc.Amount < 0.001 ||
			subacc.Currency == goex.XBT {
			continue
		}
		if subacc.Currency == fiat {
			value = subacc.Amount
			exData.addHoldings(fiat.String(), 0, subacc.Amount, value, vMap)
		} else {
			currName := estimateName(subacc.Currency.String())
			pair := getPair(currName, fiat)
			price := getPrice(pair)
			value = subacc.Amount * price.Last
			exData.addHoldings(currName, subacc.Amount, price, value, vMap)
		}
		exData.sum += value
	}
}

// printHoldingsTable() creates a nice looking table that will have data of the user's balance as well as extra calculation
func printHoldingsTable(fiat goex.Currency, exData exchangeData) {
	// sort.Sort(matrix(exData.items))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Amount",
		fmt.Sprintf("Price (%s)", fiat),
		fmt.Sprintf("Value (%s)", fiat),
		"Start (%)",
		"Last (%)"},
	)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT})

	table.SetFooter([]string{
		fmt.Sprintf("∑ %d", len(exData.items)),
		" ",
		" ",
		fmt.Sprintf("%.4f", exData.sum),
		" ",
		" "})

	table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)

	table.SetBorder(false)
	table.AppendBulk(exData.toRows())
	table.Render()
}

func printOrdersTable(odMap openOrdersMap) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Amount",
		"Value",
		"Target",
		"OrderID"},
	)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_DEFAULT})

	table.SetFooter([]string{
		fmt.Sprintf("∑ %d", len(odMap)),
		" ",
		" ",
		" ",
		" "})

	table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)

	table.SetBorder(false)
	table.AppendBulk(odMap.toRows())
	table.Render()
}

// Read the current price of a given currency pair like ETHEUR
func getPrice(pair goex.CurrencyPair) *goex.Ticker {
	price, err := apiGoex.GetTicker(pair)

	if err != nil {
		log.Print(err)
		log.Printf(" %s", pair)
	}
	return price
}

// Given a cryptocurrency name and a fiat name, getPair() will return an object that can be used in many different functions
func getPair(currName string, fiat goex.Currency) goex.CurrencyPair {
	var curr = goex.Currency{Symbol: currName, Desc: ""}
	var pair = goex.CurrencyPair{CurrencyA: curr, CurrencyB: fiat}
	return pair
}

// addHoldings() will extend the data take from the balance with a new record
func (exData *exchangeData) addHoldings(currName string, amount float64, price interface{}, value float64, vMap valuesMap) {
	var priceF float64
	var changeSinceLast []string

	vMap.storeValues(currName, value)
	oldValue := vMap[currName][0]
	changeSinceStart := getPercentage(value, oldValue, currName)

	if len(vMap[currName]) > 1 {
		recentValue := vMap[currName][len(vMap[currName])-2]
		changeSinceLast = getPercentage(value, recentValue, currName)
	}

	switch v := price.(type) {
	case float64:
		priceF = v
	case *goex.Ticker:
		priceF = v.Last
	default:
		fmt.Println("don't know the type")
	}

	var item dataItem
	item.currName = currName
	item.amount = amount
	item.price = priceF
	item.value = value
	item.changeSinceStart = strings.Join(changeSinceStart, "")
	item.changeSinceLast = strings.Join(changeSinceLast, "")
	exData.items = append(exData.items, item)

}

// calculate percentage of the change in price of a cryptocurrency
func getPercentage(value float64, oldValue float64, currName string) []string {
	change, setColor := calcChange(value, oldValue, currName)
	changeSinceValue := colorize(setColor, strconv.FormatFloat(change, 'f', 2, 64))
	return changeSinceValue
}

// calculate the change in price of a cryptocurrency and determine if it is a positive or negative change. Set the color accordingly.
func calcChange(value float64, oldValue float64, currName string) (float64, string) {
	change := (value/oldValue - 1) * 100
	setColor := "green"
	if change < 0.0 {
		setColor = "red"
	} else if change == 0.0 {
		setColor = "default"
	}
	return change, setColor
}

// connect to the api amd get the object to do further requests with
func getAPIHandle() goex.API {
	credentials := getCredentials()

	apiGoex := builder.DefaultAPIBuilder.
		APIKey(credentials.APIKey).
		APISecretkey(credentials.APISecretkey).
		ApiPassphrase(credentials.APIPassphrase).
		Build(goex.KRAKEN)
	return apiGoex
}

// set green or red for the given string
func colorize(color string, content string) []string {
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	var setting string
	switch color {
	case "red":
		setting = string(colorRed)
	case "green":
		setting = string(colorGreen)
	default:
		setting = string(colorReset)
	}
	return []string{setting, content, string(colorReset)}
}

// add data to the map that holds all values from the current session
func (vMap valuesMap) storeValues(currName string, value float64) {
	_, ok := vMap[currName]
	if !ok {
		vMap[currName] = []float64{value}
	} else {
		vMap[currName] = append(vMap[currName], value)
	}
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

// extract only the last price that was recorded already of a given cryptocurrency
func getLastValue(currName string, vMap valuesMap) float64 {
	_, ok := vMap[currName]
	if ok {
		return vMap[currName][len(vMap[currName])-1]
	}
	return 0
}

// extract all the prices that were recorded already of a given cryptocurrency
func getValues(currName string, vMap valuesMap) []float64 {
	_, ok := vMap[currName]
	if ok {
		return vMap[currName]
	}
	return []float64{0.0}
}

// read credentials to access the exchange api from disk
func getCredentials() credentials {
	var creds credentials
	credFile, _ := os.Open("credentials.txt")
	r := csv.NewReader(credFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(record) == 2 {
			switch record[0] {
			case "APIKey":
				creds.APIKey = record[1]
			case "APISecretkey":
				creds.APISecretkey = record[1]
			case "APIPassphrase":
				creds.APIPassphrase = record[1]
			}
		}
	}
	return creds
}

// write out a csv style log file that can be used to do further computation on, like in spreadsheet software
func writeStats(ts string, vmap valuesMap, statsFileLocation string) {
	// content will look like this:
	// Timestamp					ATOM	LTC	 	XLM
	// 2021-02-08T19:08:07+01:00	  10	 20		  5
	// 2021-02-08T19:09:09+01:00	  52	 15		  6

	_, err := os.Stat(statsFileLocation)
	statsFile, _ := os.OpenFile(statsFileLocation, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	var line []string
	var header []string
	writer := csv.NewWriter(statsFile)
	keys := make([]string, 0, len(vmap))
	for k := range vmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// if the file does not exist, create the header, else extend if needed
	if os.IsNotExist(err) {
		header = append([]string{"Timestamp"}, keys...)
		writer.Write(header)
		writer.Flush()
	} else {
		statsFileReader, err := os.Open(statsFileLocation)
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(statsFileReader)
		for scanner.Scan() {
			// find last header
			if strings.Contains(scanner.Text(), "Timestamp") {
				header = strings.Split(scanner.Text(), ",")
			}

		}
		defer statsFileReader.Close()
	}

	// write new header if the previous one looks different, like when a coin was sold completly for example
	if strings.Join(header[1:], ",") != strings.Join(keys, ",") {
		newHeader := append([]string{"Timestamp"}, keys...)
		writer.Write([]string{})
		writer.Write(newHeader)
		writer.Flush()
	}

	line = append(line, ts)
	for _, field := range keys {
		line = append(line, strconv.FormatFloat(getLastValue(field, vmap), 'f', 2, 64))
	}
	writer.Write(line)
	writer.Flush()
	defer statsFile.Close()
}

func generatCliArgs(currency *string, frequency *int, statsFileLocation *string, once *bool, dontWriteLog *bool, showOrderMap *bool, dontShowHoldingsMap *bool, newOrder *bool, cancelOrder *bool) {
	var version bool

	flag.StringVar(currency, "cur", "EUR", "Specify the FIAT currency to take as a baseline.")
	flag.IntVar(frequency, "freq", 360, "Specify the frequency in seconds how often the exchange API shpuld be contacted and print print the table.")
	flag.StringVar(statsFileLocation, "stats", "~/stats.txt", "Specify the location where the stats log file should be written to.")
	flag.BoolVar(once, "once", false, "Specify if the application should NOT keep running and give a new update based on the frequency but run just once and quit. Frequency setting will be ignored.")
	flag.BoolVar(showOrderMap, "orders", false, "Specify if the application should print the table of open orders.")
	flag.BoolVar(dontShowHoldingsMap, "noholdings", false, "Specify if the application should NOT print the table of holdings.")
	flag.BoolVar(newOrder, "neworder", false, "Test.")
	flag.BoolVar(cancelOrder, "cancelorder", false, "Test.")
	flag.BoolVar(&version, "version", false, "Specify if the application should print the version and quit.")
	flag.BoolVar(dontWriteLog, "nolog", false, "Specify if the application should NOT write out the stats log file.")

	flag.Parse()

	if version {
		getVersion()
		os.Exit(0)
	}
}

func main() {

	// will be filled with cli parameter
	var currency string
	var frequency int
	var once bool
	var dontWriteLog bool
	var statsFileLocation string
	var showOrderMap bool
	var dontShowHoldingsMap bool
	var newOrder bool
	var cancelOrder bool

	generatCliArgs(&currency, &frequency, &statsFileLocation, &once, &dontWriteLog, &showOrderMap, &dontShowHoldingsMap, &newOrder, &cancelOrder)

	var fiat = goex.Currency{Symbol: currency, Desc: ""}

	if newOrder {
		prompt := promptui.Select{
			Label: "What's the type of order?",
			Items: []string{"LimitBuy", "LimitSell", "MarketBuy", "MarketSell"},
		}
		_, orderType, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		var firstCurr string
		var secondCurr string
		var amount string
		var price string
		fmt.Println("Which currency to buy?")
		fmt.Scanln(&firstCurr)
		fmt.Println("Which currency to spent?")
		fmt.Scanln(&secondCurr)
		fmt.Println("What amount?")
		fmt.Scanln(&amount)
		fmt.Println("At which price?")
		fmt.Scanln(&price)

		curr1 := goex.Currency{Symbol: firstCurr}
		curr2 := goex.Currency{Symbol: secondCurr}
		pair := goex.CurrencyPair{CurrencyA: curr1, CurrencyB: curr2}
		createOrder(pair, amount, price, orderType)
		os.Exit(0)
	}

	if cancelOrder {
		var orderMap = make(openOrdersMap)
		orderMap.getOpenOrders(fiat)
		printOrdersTable(*&orderMap)

		var orderID string
		fmt.Println("What's the order ID?")
		fmt.Scanln(&orderID)
		prompt := promptui.Prompt{
			Label:     "Delete order?",
			IsConfirm: true,
		}

		ok, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		if ok == "true" {
			deletelOrder(orderID)
		}
		os.Exit(0)
	}

	var vMap = make(valuesMap)

	for {
		log.Println(fmt.Sprintf("Getting new data from %s", apiGoex.GetExchangeName()))
		acc, err := apiGoex.GetAccount()
		retryOnError(err)

		ts := time.Now().Format(time.RFC3339)
		var exHoldings = new(exchangeData)
		exHoldings.fiat = fiat
		exHoldings.extractHoldings(*acc, fiat, vMap, ts)

		if !dontWriteLog {
			go writeStats(ts, vMap, statsFileLocation)
		}

		if !dontShowHoldingsMap {
			printHoldingsTable(fiat, *exHoldings)
			fmt.Println()
		}
		if showOrderMap {
			var orderMap = make(openOrdersMap)
			orderMap.getOpenOrders(fiat)
			printOrdersTable(*&orderMap)
		}
		if once {
			break
		}

		if frequency < 60 {
			log.Println(fmt.Sprintf("Frequency is too low. It will be set to 60seconds."))
			frequency = 60
		}

		time.Sleep(time.Duration(frequency) * time.Second)
		fmt.Println()

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
