package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	goex "github.com/nntaoli-project/goex"
	"github.com/nntaoli-project/goex/builder"
	"github.com/olekukonko/tablewriter"
)

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

type matrix [][]string

func (s matrix) Len() int {
	return len(s)
}
func (s matrix) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s matrix) Less(i, j int) bool {
	return s[i][0] < s[j][0]
}

func main() {
	var currency string
	var once bool

	// flags declaration using flag package
	flag.StringVar(&currency, "c", "EUR", "Specify the FIAT currency to take as a baseline. Default is euro")
	flag.BoolVar(&once, "o", false, "Specify if the application should not keep running and give a new update every minute but run just once and quit. Default is false.")

	flag.Parse() // after declaring flags we need to call it

	apiGoex := getAPIHandle()
	var FIAT = goex.Currency{Symbol: currency, Desc: ""}

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
			value, holdings = extractHoldings(subacc, FIAT, value, holdings, apiGoex)
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

func extractHoldings(subacc goex.SubAccount, FIAT goex.Currency, value float64, holdings [][]string, apiGoex goex.API) (float64, [][]string) {
	if subacc.Currency == FIAT {
		value = subacc.Amount
		holdings = addHoldings(holdings, FIAT.String(), 0, subacc.Amount, value)
	} else {
		currName := estimateName(subacc.Currency.String())
		pair := getPair(currName, FIAT)
		price := getPrice(apiGoex, pair)
		value = subacc.Amount * price.Last
		holdings = addHoldings(holdings, currName, subacc.Amount, price, value)
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
		fmt.Sprintf("Value (%s)", FIAT)})

	table.SetFooter([]string{
		fmt.Sprintf("∑ %d", len(holdings)),
		" ",
		" ",
		fmt.Sprintf("%.2f", sum)})
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

func addHoldings(holdings [][]string, currName string, amount float64, price interface{}, value float64) [][]string {
	var priceF float64

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
			fmt.Sprintf("%.2f", value)})
	return holdings
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

type credentials struct {
	APIKey        string ""
	APISecretkey  string ""
	APIPassphrase string ""
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
