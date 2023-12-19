package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/AnnaCarter465/crypto-alert/line"
	"github.com/AnnaCarter465/crypto-alert/okx"
	"github.com/AnnaCarter465/crypto-alert/utility"
)

type RsiPerCoinPair struct {
	Pair string  `json:"pair"`
	Rsi  float64 `json:"rsi"`
}

func getOverBuyCoins() string {
	listSupportCoin, err := okx.GetListSupportCoin()
	if err != nil {
		log.Println(err)
		return "⚠️ something wrong ⚠️"
	}

	coins := listSupportCoin.Data.Contract

	log.Println("all contract coins support:", len(coins), "coins")

	if len(coins) == 0 {
		return "❌ no contract coins support ❌"
	}

	var overBuyCoins []RsiPerCoinPair

	bandwidth := make(chan struct{}, 20)

	var (
		errMux         sync.Mutex
		wg             sync.WaitGroup
		candlestickErr error
	)

	wg.Add(len(coins))

	// get candlestick by coins and get rsi > 70 ----------------------------------------------------
	log.Println("---> calculating...")
	for index, data := range coins {
		go func(index int, data string) {
			defer wg.Done()

			bandwidth <- struct{}{}
			defer func() {
				time.Sleep(time.Second * 2)
				<-bandwidth
			}()

			errMux.Lock()
			hasError := candlestickErr != nil
			errMux.Unlock()

			if hasError {
				return
			}

			pair := data + "-USDT"
			dataCandles, err := okx.GetIndexCandleStick(pair, "4H", 14)
			if err != nil {
				errMux.Lock()
				candlestickErr = err
				errMux.Unlock()
				return
			}

			rsi := utility.CalRsi(dataCandles.Data)

			if rsi > 70 {
				element := RsiPerCoinPair{Pair: pair, Rsi: rsi}
				overBuyCoins = append(overBuyCoins, element)
			}
		}(index, data)
	}

	wg.Wait()
	close(bandwidth)

	if candlestickErr != nil {
		log.Println(candlestickErr)
		return "⚠️ something wrong ⚠️"
	}

	if len(overBuyCoins) == 0 {
		return "💸 no coins rsi > 70 💸"
	}

	log.Println("over buy coins:", len(overBuyCoins), "coins")

	// sort by rsi ----------------------------------------------------
	sort.Slice(overBuyCoins, func(i, j int) bool {
		return overBuyCoins[i].Rsi > overBuyCoins[j].Rsi
	})

	var finalCoins []RsiPerCoinPair

	// get just 10 pairs ------------------------------------------------
	for i, obj := range overBuyCoins {
		if i > 10 {
			break
		}

		finalCoins = append(finalCoins, obj)
	}

	response := "\n💰 RSI over 70 💰"

	for i, v := range finalCoins {
		response += fmt.Sprintf("\n%d. %s: %f", i+1, v.Pair, v.Rsi)
	}

	return response
}

func run() {
	// Start the loop every 4
	for {
		go line.NotiToLine(getOverBuyCoins())
		time.Sleep(time.Hour * 4)
	}
}

func main() {
	now := time.Now()

	hr := now.Hour()

	var startHr int

	if now.Minute() == 0 && now.Second() == 0 {
		if hr == 7 || hr == 11 || hr == 15 || hr == 19 || hr == 23 || hr == 3 {
			run()
			return
		}
	}

	if hr > 7 && hr <= 11 {
		startHr = 11
	} else if hr > 11 && hr <= 15 {
		startHr = 15
	} else if hr > 15 && hr <= 19 {
		startHr = 19
	} else if hr > 19 && hr <= 23 {
		startHr = 23
	} else if hr > 23 || hr <= 3 {
		startHr = 3
	} else if hr > 3 && hr <= 7 {
		startHr = 7
	} else {
		panic("err")
	}

	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHr, 0, 0, 0, time.Local)

	log.Println("next startTime", startTime)

	time.Sleep(startTime.Sub(now))

	run()
}
