package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/AnnaCarter465/crypto-alert/okx"
)

type RsiPerCoinPair struct {
	Pair string  `json:"pair"`
	Rsi  float64 `json:"rsi"`
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func calRsi(data [][6]string) float64 {
	totalGain := 0.0
	totalLoss := 0.0
	periods := 14.0

	for i := 1; i < len(data); i++ {
		previous := data[i][4]
		current := data[i-1][4]

		previousClose, _ := strconv.ParseFloat(previous, 64)
		currentClose, _ := strconv.ParseFloat(current, 64)

		difference := currentClose - previousClose
		if difference >= 0 {
			totalGain += difference
		} else {
			totalLoss -= difference
		}
	}

	rs := (totalGain / periods) / (totalLoss / periods)
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		listSupportCoin, err := okx.GetListSupportCoin()
		if err != nil {
			panic(err)
		}
		coins := listSupportCoin.Data.Contract

		log.Println("contract coins support:", len(coins), "coins")

		if len(coins) == 0 {
			res := Response{Status: "success", Msg: "no contract coins support"}

			jsonBytes, err := json.Marshal(res)
			if err != nil {
				panic(err)
			}

			fmt.Fprintln(w, string(jsonBytes))
			return
		}

		var overBuyCoins []RsiPerCoinPair

		bandwidth := make(chan struct{}, 20)

		var wg sync.WaitGroup
		wg.Add(len(coins))

		// get candlestick by coins and get rsi > 70 ----------------------------------------------------
		for index, data := range coins {
			go func(index int, data string) {
				defer wg.Done()

				bandwidth <- struct{}{}
				defer func() {
					time.Sleep(time.Second * 2)
					<-bandwidth
				}()

				pair := data + "-USDT"
				dataCandles, err := okx.GetIndexCandleStick(pair, "4H", 14)
				if err != nil {
					panic(err)
				}

				rsi := calRsi(dataCandles.Data)

				if rsi > 100 {
					element := RsiPerCoinPair{Pair: pair, Rsi: rsi}
					overBuyCoins = append(overBuyCoins, element)
				}
			}(index, data)
		}

		wg.Wait()
		close(bandwidth)

		if len(overBuyCoins) == 0 {
			res := Response{Status: "success", Msg: "no coins rsi > 70"}

			jsonBytes, err := json.Marshal(res)
			if err != nil {
				panic(err)
			}

			fmt.Fprintln(w, string(jsonBytes))
			return
		}

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

		jsonBytes, err := json.Marshal(finalCoins)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, string(jsonBytes))
	})

	http.ListenAndServe(":3000", nil)
}
