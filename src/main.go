package main

import (
	"fileio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const currencySymbol = "$"
const taxRate = 0.25

func main() {
	fileContents, err := fileio.ReadLinesFromFile("resources/trades.csv")
	if err != nil {
		log.Fatal(err)
	}

	tax := calculateTax(fileContents)
	fmt.Println("~~~~~~~ tax: ", tax)
}

func calculateTax(trades []string) string {
	log := newGlobalTradeLog(trades)

	return formatCurrency(log.tax(), currencySymbol)
}

type trade struct {
	Date      string
	Symbol    string
	IsBuy     bool
	Quantity  uint
	Price     float64
	Remaining uint
}

// companyTradeLog is all trades for a common symbol
type companyTradeLog struct {
	Trades []*trade
}

// globalTradeLog is all trades, sharded by company symbol
type globalTradeLog map[string]*companyTradeLog

func (t globalTradeLog) tax() float64 {
	var tax float64

	// use channel with WaitGroup to allow (but not guarantee) parallelism
	// companyTradeLog.tax() is responsible for putting an amount on the chan and calling waitGroup.Done()
	taxes := make(chan float64, len(t))
	waitGroup := &sync.WaitGroup{}

	// total tax due is the sum of the tax due for each company
	// start each company's calculation in a parallel goroutine
	waitGroup.Add(len(t))
	for _, companyLog := range t {
		go companyLog.tax(waitGroup, taxes)
	}

	// block until each company's goroutine calls Done() on the WaitGroup
	waitGroup.Wait()
	close(taxes)

	for companyTax := range taxes {
		tax += companyTax
	}

	return tax
}

// tax iterates in order over all Buys in this log and find computes the tax based on a static 25% rate
// mutates this companyTradeLog in place (specifically, modifies trade.Remaining to compute cost FIFO cost basis)
func (t *companyTradeLog) tax(waitGroup *sync.WaitGroup, taxes chan<- float64) {
	defer waitGroup.Done()

	var profit float64

	for _, sale := range t.Trades {
		if sale.IsBuy {
			// this trade is a buy, no tax calculation to do, skip
			continue
		}

		proceeds := sale.Price * float64(sale.Quantity)

		// we need to find our Quantity of shares somewhere ... combination of past Buys and/or future Buys
		var basis float64
		needToFind := sale.Quantity

		// iterate FIFO over all our buys, including buys after this sale if necessary (short sales supported)
		for _, buy := range t.Trades {
			if !buy.IsBuy {
				// this trade is a sale, we're looking for Buys, skip
				continue
			}

			if buy.Remaining < 1 {
				// this lot has been all sold off already, skip
				continue
			}

			if buy.Remaining >= needToFind {
				// this is the last lot (not necessarily first/only lot) we need to check for this sale
				basis += buy.Price * float64(needToFind)

				// deduct from this lot for future calculations, then we're done computing this basis
				buy.Remaining -= needToFind
				break
			}

			// buy.Remaining < needToFind
			basis += buy.Price * float64(buy.Remaining)

			// don't break, keep going, deduct from this lot's remaining quantity AND deduct from needToFind
			needToFind -= buy.Remaining
			buy.Remaining = 0
		}

		profit += proceeds - basis
	}

	// no tax carry-forward accounting if company lost money in calendar year
	if profit <= 0 {
		taxes <- 0
	} else {
		taxes <- profit * taxRate
	}

	return
}

func formatCurrency(amount float64, symbol string) string {
	// exactly 2 decimal places
	formatted := strconv.FormatFloat(amount, 'f', 2, 64)

	// parentheses for negative values. 0.00 doesn't get parentheses
	if amount < 0 {
		formatted = fmt.Sprintf("(%s)", formatted)
	}

	return fmt.Sprintf("%s%s", symbol, formatted)
}

func parseTrade(tradeLine string) (*trade, error) {
	tokens := strings.Split(tradeLine, ",")
	if len(tokens) != 5 {
		return nil, errors.New("given string does not contain the required 5 fields")
	}

	isBuy := strings.ToLower(tokens[2]) == "b"

	quantity, err := strconv.ParseUint(tokens[3], 10, strconv.IntSize)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing quantity of %s (must ne non-negative integer)", tokens[3])
	}

	price, err := strconv.ParseFloat(tokens[4], 64)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing trade price of %s", tokens[4])
	}

	var remaining uint
	if isBuy {
		remaining = uint(quantity)
	}

	return &trade{
		Date:      tokens[0],
		Symbol:    tokens[1],
		IsBuy:     isBuy,
		Quantity:  uint(quantity),
		Price:     price,
		Remaining: remaining,
	}, nil
}

func newGlobalTradeLog(tradeLines []string) globalTradeLog {
	log := globalTradeLog{}

	for _, tradeLine := range tradeLines {
		trd, err := parseTrade(tradeLine)
		if err != nil {
			// swallow this error and continue parsing
			fmt.Println("error parsing trade from log, skipping: ", err.Error())
			continue
		}

		symbol := strings.ToLower(trd.Symbol)
		if companyLog, ok := log[symbol]; ok {
			// we already know about this company, append to its companyTradeLog
			companyLog.Trades = append(companyLog.Trades, trd)
		} else {
			// first trade we've encountered for this company, initialize a companyTradeLog
			log[symbol] = &companyTradeLog{
				Trades: []*trade{trd},
			}
		}
	}

	return log
}
