package main

import (
	"fileio"

	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const currencySymbol = "$"
const pollInterval = 100 * time.Millisecond
const taxRate = 0.25

func main() {
	trades, err := fileio.ReadLinesFromFile("resources/trades.csv")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("calculating total tax using wait loop: ", calculateTax(trades))

	fmt.Println("re-calculating total tax using sync.WaitGroup (should match): ", calculateTaxGroup(trades))
}

func calculateTax(trades []string) string {
	var log globalTradeLog = newGlobalTradeLog(trades)

	return formatCurrency(log.tax(), currencySymbol)
}

func calculateTaxGroup(trades []string) string {
	var log globalTradeLog = newGlobalGroupTradeLog(trades)

	return formatCurrency(log.tax(), currencySymbol)
}

// globalTradeLog is the high level interface providing tax calculation across all companies
type globalTradeLog interface {
	// tax returns the net tax due across all companies
	tax() float64
}

type trade struct {
	Date      string
	Symbol    string
	IsBuy     bool
	Quantity  uint
	Price     float64
	Remaining uint
}

// companyTradeLog is all trades for one company
// low level type used by various globalTradeLog implementations
type companyTradeLog struct {
	Trades []*trade
}

// waitLoopGlobalTradeLog is a globalTradeLog implementation that uses a polling loop to wait for each company's goroutine to complete
type waitLoopGlobalTradeLog struct {
	companyTradeLogs map[string]*companyTradeLog
}

// waitLoopGlobalTradeLog is a globalTradeLog implementation that uses sync.WaitGroup to wait for each company's goroutine to return
type waitGroupGlobalTradeLog struct {
	companyTradeLogs map[string]*companyTradeLog
}

func (t *waitLoopGlobalTradeLog) tax() float64 {
	taxes := make(chan float64, len(t.companyTradeLogs))

	// total tax due is the sum of the tax due for each company
	// start each company's calculation in a goroutine
	// like the WaitGroup approach, this allows but not guarantee parallelism
	// companyTradeLog.tax() is responsible for putting an amount on the chan then returning -- no waitGroup.Done() or other explicit "done" notification
	for _, companyLog := range t.companyTradeLogs {
		go companyLog.tax(taxes, nil)
	}

	var tax float64
	var counter int
WAITLOOP:
	for true {
		select {
		case partialTax := <-taxes:
			// accumulate the total tax, and track how many companies have completed
			tax += partialTax
			counter++

			// if all companies are done calculating
			if counter >= len(t.companyTradeLogs) {
				break WAITLOOP
			}
		default:
			// at least one company is still being calculated
			// this polling is avoidable using the WaitGroup approach
			time.Sleep(pollInterval)
		}
	}

	return tax
}

func (t *waitGroupGlobalTradeLog) tax() float64 {
	taxes := make(chan float64, len(t.companyTradeLogs))
	waitGroup := &sync.WaitGroup{}

	// total tax due is the sum of the tax due for each company
	// start each company's calculation in a goroutine
	// like the wait loop approach, this allows but not guarantee parallelism
	// companyTradeLog.tax() is responsible for putting an amount on the chan and calling waitGroup.Done() before returning
	waitGroup.Add(len(t.companyTradeLogs))
	for _, companyLog := range t.companyTradeLogs {
		go companyLog.tax(taxes, waitGroup)
	}

	// block until each company's goroutine calls Done() on the WaitGroup
	waitGroup.Wait()
	close(taxes)

	var tax float64
	for companyTax := range taxes {
		tax += companyTax
	}

	return tax
}

// tax iterates in order over all Buys in this log and find computes the tax based on a static 25% rate
// mutates this companyTradeLog in place (specifically, modifies trade.Remaining to compute cost FIFO cost basis)
func (t *companyTradeLog) tax(taxes chan<- float64, waitGroup *sync.WaitGroup) {
	if waitGroup != nil {
		defer waitGroup.Done()
	}

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

func newGlobalTradeLog(tradeLines []string) *waitLoopGlobalTradeLog {
	return &waitLoopGlobalTradeLog{
		companyTradeLogs: buildCompanyTradeLogMap(tradeLines),
	}
}

func newGlobalGroupTradeLog(tradeLines []string) *waitGroupGlobalTradeLog {
	return &waitGroupGlobalTradeLog{
		companyTradeLogs: buildCompanyTradeLogMap(tradeLines),
	}
}

func buildCompanyTradeLogMap(tradeLines []string) map[string]*companyTradeLog {
	companyTradeLogs := map[string]*companyTradeLog{}

	for _, tradeLine := range tradeLines {
		trd, err := parseTrade(tradeLine)
		if err != nil {
			// swallow this error and continue parsing
			fmt.Println("error parsing trade from log, skipping: ", err.Error())
			continue
		}

		symbol := strings.ToLower(trd.Symbol)
		if companyLog, ok := companyTradeLogs[symbol]; ok {
			// we already know about this company, append to its companyTradeLog
			companyLog.append(trd)
		} else {
			// first trade we've encountered for this company, initialize a companyTradeLog
			companyTradeLogs[symbol] = &companyTradeLog{
				Trades: []*trade{trd},
			}
		}
	}

	return companyTradeLogs
}

func (t *companyTradeLog) append(trd *trade) {
	t.Trades = append(t.Trades, trd)
}
