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

// extractHoldings() will pick out various parts of the user's balance data write a human-readable string
func extractHoldings(subacc goex.SubAccount, FIAT goex.Currency, value float64, holdings [][]string, vMap valuesMap, ts string, apiGoex goex.API) (float64, [][]string) {
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

// printTable() creates a nice looking table that will have data of the user's balance as well as extra calculation
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

// Read the current price of a given currency pair like ETHEUR
func getPrice(apiGoex goex.API, pair goex.CurrencyPair) *goex.Ticker {
	price, err := apiGoex.GetTicker(pair)

	if err != nil {
		log.Print(err)
		log.Printf(" %s", pair)
	}
	return price
}

// Given a cryptocurrency name and a fiat name, getPair() will return an object that can be used in many different functions
func getPair(currName string, FIAT goex.Currency) goex.CurrencyPair {
	var curr = goex.Currency{Symbol: currName, Desc: ""}
	var pair = goex.CurrencyPair{CurrencyA: curr, CurrencyB: FIAT}
	return pair
}

// addHoldings() will extend the data take from the balance with a new record
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

// calculate percentage of the change in price of a cryptocurrency
func getPercentage(value float64, oldValue float64, vMap valuesMap, currName string) []string {
	change, setColor := calcChange(value, oldValue, vMap, currName)
	changeSinceValue := colorize(setColor, strconv.FormatFloat(change, 'f', 2, 64))
	return changeSinceValue
}

// calculate the change in price of a cryptocurrency and determine if it is a positive or negative change. Set the color accordingly.
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
func storeValues(currName string, value float64, vMap valuesMap) valuesMap {
	_, ok := vMap[currName]
	if !ok {
		vMap[currName] = []float64{value}
	} else {
		vMap[currName] = append(vMap[currName], value)
	}
	return vMap
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

// write out a csv style log file that can be used to do further computation on, like in spreadsheet software
func writeStats(ts string, vmap valuesMap, statsFileLocation string) {
	//// content will look like this
	// Timestamp					LTC	  XLM	ATOM
	// 2021-02-08T19:08:07+01:00	10	  20	5
	// 2021-02-08T19:09:09+01:00	52	  15	6

	_, err := os.Stat(statsFileLocation)
	statsFile, _ := os.OpenFile(statsFileLocation, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	var line []string
	writer := csv.NewWriter(statsFile)
	var header []string

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
	var frequency int
	var once bool
	var version bool
	var dontWriteLog bool
	var statsFileLocation string

	// flags declaration using flag package
	flag.StringVar(&currency, "c", "EUR", "Specify the FIAT currency to take as a baseline.")
	flag.IntVar(&frequency, "f", 360, "Specify the frequency in seconds how often the exchange API shpuld be contacted and print print the table.")
	flag.StringVar(&statsFileLocation, "stats", "~/stats.txt", "Specify the location where the stats log file should be written to.")
	flag.BoolVar(&once, "o", false, "Specify if the application should NOT keep running and give a new update based on the frequency but run just once and quit. Frequency setting will be ignored.")
	flag.BoolVar(&version, "v", false, "Specify if the application should print the version and quit.")
	flag.BoolVar(&dontWriteLog, "nolog", false, "Specify if the application should NOT write out the stats log file.")

	flag.Parse() // after declaring flags we need to call it
	if version {
		getVersion()
		return
	}

	apiGoex := getAPIHandle()
	var FIAT = goex.Currency{Symbol: currency, Desc: ""}
	var vMap = make(valuesMap)
	errorCounter := 0

	for {
		log.Println(fmt.Sprintf("Getting new data from %s", apiGoex.GetExchangeName()))
		acc, err := apiGoex.GetAccount()
		if err != nil {
			// TODO: Needs to be made more granular but let's  assume that this error is just a temporary
			// 4xx or 5xx. A retry with a delay will hopefully work. Give up after 5 consecutive errors,
			errorCounter++

			fmt.Print(err)
			if errorCounter > 5 {
				return
			}

			time.Sleep(60 * time.Second)
			continue
		}
		errorCounter = 0

		var holdings [][]string
		var sum float64
		var value float64

		ts := time.Now().Format(time.RFC3339)

		for _, subacc := range acc.SubAccounts {
			// skip small amounts, skip euro amount, skip redundant XBT as BTC is already reported for Bitcoin
			if subacc.Amount < 0.001 ||
				subacc.Currency == goex.XBT {
				continue
			}

			value, holdings = extractHoldings(subacc, FIAT, value, holdings, vMap, ts, apiGoex)
			sum += value
		}

		if !dontWriteLog {
			go writeStats(ts, vMap, statsFileLocation)
		}

		printTable(FIAT, holdings, sum)

		if once {
			break
		}

		if frequency < 60 {
			fmt.Println("Frequency is too low. It will be set to 60seconds.")
			frequency = 60
		}

		time.Sleep(time.Duration(frequency) * time.Second)
		fmt.Println()
	}
}
