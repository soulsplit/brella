# brella

This tool aims to provide a command line oriented cryptocurrency exchange frontend.

## Supported

- kraken.com

## Dependencies

It is using the geox library to interact with the exchanges:  
[github.com/nntaoli-project/goex](github.com/nntaoli-project/goex)  
However this library was missing an order related setting which is why a patched local fork is used for now.  
[github.com/soulsplit/goex](github.com/soulsplit/goex)

To print the table the tablewriter library is used:  
[github.com/olekukonko/tablewriter](github.com/olekukonko/tablewriter)

## Compile

``` shell
go build ./...
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

## Holdings Table Sample Output

``` shell
  NAME |  AMOUNT  | PRICE (EUR) | VALUE (EUR)  | START (%) | LAST (%)  
-------+----------+-------------+--------------+-----------+-----------
  BTC  |   0.5000 |    32000.00 |    16000.00  |      0.56 |     0.11
  ETH  |   1.0000 |     1400.30 |     1400.30  |      0.47 |    -0.31
  EUR  |   0.0000 |        1.00 |        1.00  |      0.00 |     0.00 
-------+----------+-------------+--------------+-----------+-----------
  ∑ 3  |          |             |    17401.30  |           |           

```

## Orders Table Sample Output

``` shell
   NAME   | AMOUNT |   VALUE  |       ORDERID        
----------+--------+----------+----------------------
  BTC_EUR |   0.25 | 40000.00 | OAJOGH-62PAC-IST3D4  
  ETH_EUR |   0.10 |  1800.00 | OA76XK-VLZ3C-WD4LBC  
----------+--------+----------+----------------------
      ∑ 2 |        |          |                      
----------+--------+----------+----------------------
```

The request will be rerun every 6 minutes. Use `CTRL+C` to quit.

## Following Parameters are supported

``` shell
./brella -h
Usage of ./brella:
  -cur string
    	Specify the FIAT currency to take as a baseline. (default "EUR")
  -freq int
    	Specify the frequency in seconds how often the exchange API shpuld be contacted and print print the table. (default 360)
  -noholdings
    	Specify if the application should NOT print the table of holdings.
  -nolog
    	Specify if the application should NOT write out the stats log file.
  -once
    	Specify if the application should NOT keep running and give a new update based on the frequency but run just once and quit. Frequency setting will be ignored.
  -orders
    	Specify if the application should print the table of open orders.
  -stats string
    	Specify the location where the stats log file should be written to. (default "~/stats.txt")
  -version
    	Specify if the application should print the version and quit.
  ```
