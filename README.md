## what is this?
A simple capital gains tax calculator. Computes the cost basis and capital gains tax, including short sales and split lots. Applies a simple 25% flat tax rate. Does not support short vs long terms, marginal tax brackets, or loss carryforward.


## what is this really?
A demo of two approaches for dispatching and waiting for parallel workers. In either approach, one goroutine is dispatched for each company stock symbol, which allows -- but don't guarantee -- parallelism.
1) wait loop that uses the chan both for passing values _and_ implicitly tracking goroutine completion
1) using Go's sync.WaitGroup to explicitly track goroutine completion

I didn't find examples that were sufficiently complete or explained. Particularly:
 - main/controller function spawning goroutines and waiting, rather than parallel consumer goroutines
 - invoking external functions as goroutines, rather than nested anonymous functions that share scope
 - `all goroutines are asleep - deadlock!` errors when ranging over an unclosed chan
 - blocking vs non-blocking reads


## usage

Place a trades.csv file in the resources directory with the following format (B => buy, S => sale):

||||||
|---|---|---|---|---|
|2018-01-03|AAPL|B|50|80.00|
|2018-01-05|AAPL|B|60|100.00|
|2018-02-05|AAPL|S|70|130.00|
|2018-02-08|AAPL|S|10|90.00|
|2018-02-08|NUAN|B|10|10.00|
|2018-03-10|AAPL|S|80|120.00|
|2018-03-12|AAPL|B|10|70.00|
|2018-02-08|NUAN|S|10|20.00|
|2018-04-08|AAPL|B|70|160.00|
|2018-07-10|NVDA|B|110|230.00|
|2018-07-22|NVDA|B|90|244.00|
|2018-10-22|NVDA|S|100|259.00|
|2018-10-26|NVDA|S|50|230.00|
|2018-11-25|NVDA|B|50|200.00|

```
go get "github.com/pkg/errors"

go run src/main.go
```