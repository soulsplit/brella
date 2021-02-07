package main

import (
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

	goex "github.com/nntaoli-project/goex"
	"github.com/nntaoli-project/goex/builder"
	"github.com/olekukonko/tablewriter"
)

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

type valuesMap map[string][]float64

func (s matrix) Len() int {
	return len(s)
}
func (s matrix) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s matrix) Less(i, j int) bool {
	return s[i][0] < s[j][0]
}

func getVersion() {
	fmt.Println("0.0.1")
}

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

func extractHoldings(subacc goex.SubAccount, FIAT goex.Currency, value float64, holdings [][]string, vMap valuesMap, apiGoex goex.API) (float64, [][]string) {
	if subacc.Currency == FIAT {
		value = subacc.Amount
		holdings = addHoldings(holdings, FIAT.String(), 0, subacc.Amount, value, vMap)
	} else {
		currName := estimateName(subacc.Currency.String())
		pair := getPair(currName, FIAT)
		price := getPrice(apiGoex, pair)
		value = subacc.Amount * price.Last
		holdings = addHoldings(holdings, currName, subacc.Amount, price, value, vMap)
	}
	return value, holdings
}

func printTable(FIAT goex.Currency, holdings [][]string, sum float64) {
	sort.Sort(matrix(holdings))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Amount",
		fmt.Sprintf("Price (%s)", FIAT),
		fmt.Sprintf("Value (%s)", FIAT),
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
		fmt.Sprintf("∑ %d", len(holdings)),
		" ",
		" ",
		fmt.Sprintf("%.2f", sum),
		" ",
		" "})

	table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)

	table.SetBorder(false)
	table.AppendBulk(holdings)
	table.Render()
}

func getPrice(apiGoex goex.API, pair goex.CurrencyPair) *goex.Ticker {
	price, err := apiGoex.GetTicker(pair)

	if err != nil {
		log.Print(err)
		log.Printf(" %s", pair)
	}
	return price
}

func getPair(currName string, FIAT goex.Currency) goex.CurrencyPair {
	var curr = goex.Currency{Symbol: currName, Desc: ""}
	var pair = goex.CurrencyPair{CurrencyA: curr, CurrencyB: FIAT}
	return pair
}

func addHoldings(holdings [][]string, currName string, amount float64, price interface{}, value float64, vMap valuesMap) [][]string {
	var priceF float64
	var changeSinceLast []string

	vMap = storeValues(currName, value, vMap)
	oldValue := vMap[currName][0]
	changeSinceStart := getPercentage(value, oldValue, vMap, currName)

	if len(vMap[currName]) > 1 {
		recentValue := vMap[currName][len(vMap[currName])-2]
		changeSinceLast = getPercentage(value, recentValue, vMap, currName)
	}

	switch v := price.(type) {
	case float64:
		priceF = v
	case *goex.Ticker:
		priceF = v.Last
	default:
		fmt.Println("don't know the type")
	}

	holdings = append(holdings,
		[]string{fmt.Sprintf("%s", currName),
			fmt.Sprintf("%.4f", amount),
			fmt.Sprintf("%.2f", priceF),
			fmt.Sprintf("%.2f", value),
			strings.Join(changeSinceStart, ""),
			strings.Join(changeSinceLast, "")})
	return holdings
}

func getPercentage(value float64, oldValue float64, vMap valuesMap, currName string) []string {
	change, setColor := calcChange(value, oldValue, vMap, currName)
	changeSinceValue := colorize(setColor, strconv.FormatFloat(change, 'f', 2, 64))
	return changeSinceValue
}

func calcChange(value float64, oldValue float64, vMap valuesMap, currName string) (float64, string) {
	change := (value/oldValue - 1) * 100
	setColor := "green"
	if change < 0.0 {
		setColor = "red"
	} else if change == 0.0 {
		setColor = "default"
	}
	return change, setColor
}

func getAPIHandle() goex.API {
	credentials := getCredentials()

	apiGoex := builder.DefaultAPIBuilder.
		APIKey(credentials.APIKey).
		APISecretkey(credentials.APISecretkey).
		ApiPassphrase(credentials.APIPassphrase).
		Build(goex.KRAKEN)
	return apiGoex
}

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

func storeValues(currName string, value float64, vMap valuesMap) valuesMap {
	_, ok := vMap[currName]
	if !ok {
		vMap[currName] = []float64{value}
	} else {
		vMap[currName] = append(vMap[currName], value)
	}
	return vMap
}

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
			if record[0] == "APIKey" {
				creds.APIKey = record[1]
			} else if record[0] == "APISecretkey" {
				creds.APISecretkey = record[1]
			} else if record[0] == "APIPassphrase" {
				creds.APIPassphrase = record[1]
			}
		}
	}
	return creds
}

// func getHistory(){
// historic, err := apiGoex.GetOrderHistorys(pair)

// if err != nil {
// 	log.Print(err)
// 	log.Printf(" %s", pair)
// 	continue
// }
// for _, data := range historic {
// 	log.Printf("%f", data.AvgPrice)
// }
// }

func main() {
	var currency string
	var once bool
	var version bool

	// flags declaration using flag package
	flag.StringVar(&currency, "c", "EUR", "Specify the FIAT currency to take as a baseline.")
	flag.BoolVar(&once, "o", false, "Specify if the application should not keep running and give a new update every minute but run just once and quit.")
	flag.BoolVar(&version, "v", false, "Specify if the application should print the version and quit.")

	flag.Parse() // after declaring flags we need to call it
	if version {
		getVersion()
		return
	}

	apiGoex := getAPIHandle()
	var FIAT = goex.Currency{Symbol: currency, Desc: ""}
	var vMap = make(valuesMap)

	for {
		log.Println(fmt.Sprintf("Getting new data from %s", apiGoex.GetExchangeName()))
		acc, _ := apiGoex.GetAccount()
		var holdings [][]string
		var sum float64
		var value float64

		for _, subacc := range acc.SubAccounts {
			// skip small amounts, skip euro amount, skip redundant XBT as BTC is already reported for Bitcoin
			if subacc.Amount < 0.001 ||
				subacc.Currency == goex.XBT {
				continue
			}
			value, holdings = extractHoldings(subacc, FIAT, value, holdings, vMap, apiGoex)
			sum += value
		}

		printTable(FIAT, holdings, sum)
		if once {
			break
		}
		time.Sleep(60 * time.Second)
		fmt.Println()
	}

}
