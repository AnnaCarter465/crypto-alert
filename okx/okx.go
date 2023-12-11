package okx

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/AnnaCarter465/crypto-alert/httprequest"
)

var domain = "https://www.okx.com"

type OkxResponse[T any] struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type TradingEndpoint struct {
	Contract []string `json:"contract"`
	Option   []string `json:"option"`
	Spot     []string `json:"spot"`
}

func GetListSupportCoin() (OkxResponse[TradingEndpoint], error) {
	res, err := httprequest.Request("GET", domain+"/api/v5/rubik/stat/trading-data/support-coin")
	if err != nil {
		return OkxResponse[TradingEndpoint]{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return OkxResponse[TradingEndpoint]{}, err
	}

	var response OkxResponse[TradingEndpoint]
	err = json.Unmarshal(body, &response)
	if err != nil {
		return OkxResponse[TradingEndpoint]{}, err
	}

	return response, nil
}

func GetIndexCandleStick(pair, bar string, limit int) (OkxResponse[[][6]string], error) {
	res, err := httprequest.Request("GET", domain+"/api/v5/market/index-candles?before&limit="+strconv.Itoa(limit)+"&bar="+bar+"&instId="+pair)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	var response OkxResponse[[][6]string]
	err = json.Unmarshal(body, &response)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	return response, nil
}
