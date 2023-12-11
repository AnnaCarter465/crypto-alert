package okx

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/AnnaCarter465/crypto-alert/httprequest"
)

var (
	domain        = "https://www.okx.com"
	defaultErrMsg = "something wrong with OKX API"
)

type OkxResponse[T any] struct {
	Code         interface{} `json:"code"`
	Msg          string      `json:"msg"`
	Data         T           `json:"data"`
	DetailMsg    string      `json:"detailMsg"`
	ErrorCode    string      `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
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

	if res.StatusCode != 200 {
		return OkxResponse[TradingEndpoint]{}, errors.New(defaultErrMsg)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return OkxResponse[TradingEndpoint]{}, err
	}

	var response OkxResponse[TradingEndpoint]
	err = json.Unmarshal(body, &response)
	if err != nil {
		return OkxResponse[TradingEndpoint]{}, err
	}

	if response.Code != "0" {
		errMsg := response.ErrorMessage
		if errMsg == "" {
			errMsg = defaultErrMsg
		}

		return OkxResponse[TradingEndpoint]{}, errors.New(errMsg)
	}

	return response, nil
}

func GetIndexCandleStick(pair, bar string, limit int) (OkxResponse[[][6]string], error) {
	res, err := httprequest.Request("GET", domain+"/api/v5/market/index-candles?before&limit="+strconv.Itoa(limit)+"&bar="+bar+"&instId="+pair)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return OkxResponse[[][6]string]{}, errors.New(defaultErrMsg)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	var response OkxResponse[[][6]string]
	err = json.Unmarshal(body, &response)
	if err != nil {
		return OkxResponse[[][6]string]{}, err
	}

	if response.Code != "0" {
		errMsg := response.ErrorMessage
		if errMsg == "" {
			errMsg = defaultErrMsg
		}

		return OkxResponse[[][6]string]{}, errors.New(errMsg)
	}

	return response, nil
}
