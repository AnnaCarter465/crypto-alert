package utility

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AnnaCarter465/crypto-alert/model"
)

func CalRsi(data [][6]string) float64 {
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

func ThrowInternalServerError(err error, w http.ResponseWriter) {
	res := model.Response{Status: "error", Msg: err.Error()}
	errJson, _ := json.Marshal(res)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(errJson)
}
