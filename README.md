# brella

This tool aims to provide a command line oriented cryptocurrency exchange frontend.

## Supported

- kraken.com

## Libraries

It is using the geox library to interact with the exchanges:  
[github.com/nntaoli-project/goex](github.com/nntaoli-project/goex)

To print the table the tablewriter library is used:  
[github.com/olekukonko/tablewriter](github.com/olekukonko/tablewriter)

## Compile

``` shell
go build brella.go
```

## Configure

Create a filed called `credentials.txt` with your crenditials set like in this example:

``` csv
APIKey,ABCD
APISecretkey,xyz==
APIPassphrase,Pass
```

Note: For kraken.com the `APIPassphrase` value can be left empty.

## Run

``` shell
./brella
```

## Sample Output

``` shell
2021/01/01 08:00:00 Getting new data from kraken.com
  NAME |  AMOUNT  | PRICE (EUR) | VALUE (EUR)  | START (%) | LAST (%)  
-------+----------+-------------+--------------+-----------+-----------
  BTC  |   0.5000 |    32000.00 |    16000.00  |      0.56 |     0.11
  ETH  |   1.0000 |     1400.30 |     1400.30  |      0.47 |    -0.31
  EUR  |   0.0000 |        1.00 |        1.00  |      0.00 |     0.00 
-------+----------+-------------+--------------+-----------+-----------
  ∑ 3  |          |             |    17401.30  |           |           

```

The request will be rerun every 6 minutes. Use `CTRL+C` to quit.

## Following Parameters are supported

``` shell
./brella -h
Usage of ./brella:
  -c string
        Specify the FIAT currency to take as a baseline. (default "EUR")
  -f int
        Specify the frequency in seconds how often the exchange API shpuld be contacted and print print the table. (default 360)
  -nolog
        Specify if the application should NOT write out a log.
  -o    Specify if the application should NOT keep running and give a new update based on the frequency but run just once and quit. Frequency setting will be ignored.
  -v    Specify if the application should print the version and quit.
  ```
