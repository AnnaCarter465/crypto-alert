package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/AnnaCarter465/crypto-alert/okx"
	"github.com/AnnaCarter465/crypto-alert/utility"
	"github.com/gorilla/mux"
)

type RsiPerCoinPair struct {
	Pair string  `json:"pair"`
	Rsi  float64 `json:"rsi"`
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func getOverBuyCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	listSupportCoin, err := okx.GetListSupportCoin()
	if err != nil {
		panic(err)
	}
	coins := listSupportCoin.Data.Contract

	log.Println("all contract coins support:", len(coins), "coins")

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
	log.Println("---> calculating...")
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

			rsi := utility.CalRsi(dataCandles.Data)

			if rsi > 70 {
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

	jsonBytes, err := json.Marshal(finalCoins)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, string(jsonBytes))
}

func main() {
	router := mux.NewRouter()

	subRouter := router.PathPrefix("/api/crypto-alert").Subrouter()

	subRouter.HandleFunc("/contract", getOverBuyCoins)

	http.ListenAndServe(":3000", router)
}
