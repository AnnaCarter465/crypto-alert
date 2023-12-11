package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/AnnaCarter465/crypto-alert/model"
	"github.com/AnnaCarter465/crypto-alert/okx"
	"github.com/AnnaCarter465/crypto-alert/utility"
	"github.com/gorilla/mux"
)

type RsiPerCoinPair struct {
	Pair string  `json:"pair"`
	Rsi  float64 `json:"rsi"`
}

func getOverBuyCoins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	listSupportCoin, err := okx.GetListSupportCoin()
	if err != nil {
		log.Println(err)
		utility.ThrowInternalServerError(err, w)
		return
	}

	coins := listSupportCoin.Data.Contract

	log.Println("all contract coins support:", len(coins), "coins")

	if len(coins) == 0 {
		res := model.Response{Status: "success", Msg: "no contract coins support"}
		resJson, _ := json.Marshal(res)
		w.Write(resJson)
		return
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
		utility.ThrowInternalServerError(candlestickErr, w)
		return
	}

	if len(overBuyCoins) == 0 {
		res := model.Response{Status: "success", Msg: "no coins rsi > 70"}
		resJson, _ := json.Marshal(res)
		w.Write(resJson)
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

	resJson, _ := json.Marshal(finalCoins)
	w.Write(resJson)
}

func main() {
	router := mux.NewRouter()

	subRouter := router.PathPrefix("/api/crypto-alert").Subrouter()

	subRouter.HandleFunc("/contract", getOverBuyCoins)

	http.ListenAndServe(":3000", router)
}
