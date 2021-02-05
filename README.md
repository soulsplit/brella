# brella

This tool aims to provide a command line oriented cryptocurrency exchange frontend.

## Supported

- kraken.com

## Libraries

It is using the geox library to interact with the exchanges:
github.com/nntaoli-project/goex

To print the table the tablewriter library is used:
github.com/olekukonko/tablewriter

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
  NAME |  AMOUNT  | PRICE (EUR) | VALUE (EUR)  
-------+----------+-------------+--------------
  BTC  |   0.5000 |    32000.00 |    16000.00    
  ETH  |   1.0000 |     1400.30 |     1400.30  
  EUR  |   0.0000 |        1.00 |        1.00
-------+----------+-------------+--------------
  ∑ 3  |          |             |    17401.30  

```

The request will be rerun every minute. Use `CTRL+C` to quit.
