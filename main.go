package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type ExchangeRates struct {
	Result             string             `json:"result"`
	Provider           string             `json:"provider"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int                `json:"time_last_update_unix"`
	TimeLastUpdateUtc  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int                `json:"time_next_update_unix"`
	TimeNextUpdateUtc  string             `json:"time_next_update_utc"`
	TimeEolUnix        int                `json:"time_eol_unix"`
	BaseCode           string             `json:"base_code"`
	Rates              map[string]float64 `json:"rates"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "[amount] [source currency code] [target currency code]",
		Short: "convert amount to any currency",
		Long:  "convert amount from any currency to any other currency",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			amount, err := validateAmount(args[0])
			if err != nil {
				log.Fatal(err)
			}

			sourceCurrency, targetCurrency, err := validateCurrencyCodes(args[1], args[2])
			if err != nil {
				log.Fatal(err)
			}

			result, err := convertCurrency(amount, sourceCurrency, targetCurrency)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("%f %v is %f %v", amount, sourceCurrency, result, targetCurrency)
		},
	}

	rootCmd.Execute()
}

func validateAmount(amountInput string) (float64, error) {
	amount, err := strconv.ParseFloat(amountInput, 64)
	if err != nil {
		return 0, errors.New("Invalid amount. Please provide a valid number.")
	}

	return amount, nil
}

func validateCurrencyCodes(sourceCurrency string, targetCurrency string) (string, string, error) {
	re := regexp.MustCompile("^[a-zA-Z]{3}$")
	if !re.Match([]byte(sourceCurrency)) || !re.Match([]byte(targetCurrency)) {
		return "", "", errors.New("Invalid currency code(s). Currency codes must be 3 characters long (e.g., USD).")
	}

	return strings.ToUpper(sourceCurrency), strings.ToUpper(targetCurrency), nil
}

func convertCurrency(amount float64, from string, to string) (float64, error) {
	res, err := http.Get("https://open.er-api.com/v6/latest/" + from)
	if err != nil {
		return 0, errors.New("Failed to retrieve exchange rates. Please try again later.")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, errors.New("Failed to read response body")
	}

	var exchangeRates ExchangeRates

	err = json.Unmarshal(body, &exchangeRates)
	if err != nil {
		return 0, errors.New("Failed to unmarshal json")
	}

	if exchangeRates.Result != "success" {
		return 0, errors.New("Exchange rate api failed to get the rates")
	}

	return amount * exchangeRates.Rates[to], nil
}
