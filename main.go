package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
	"github.com/mtlynch/gofn-prosper/types"

	"github.com/mtlynch/prosperbot/account"
	"github.com/mtlynch/prosperbot/buyer"
	"github.com/mtlynch/prosperbot/notes"
)

func parseCredentials(path string) (creds types.ClientCredentials, err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return types.ClientCredentials{}, err
	}
	err = json.Unmarshal(file, &creds)
	if err != nil {
		return types.ClientCredentials{}, err
	}
	return creds, nil
}

func main() {
	log.Println("Starting up!")
	credsPath := flag.String("creds", "/opt/prosper-creds.json", "client credentials file")
	isBuyingEnabled := flag.Bool("enable-buying", false, "is listing buying enabled?")
	flag.Parse()
	creds, err := parseCredentials(*credsPath)
	if err != nil {
		log.Fatalf("failed to parse credentials: %v", err)
	}
	c := prosper.NewClient(creds)
	f := prosper.SearchFilter{
		EstimatedReturn: types.Float64Range{Min: types.CreateFloat64(0.0849)},
		ListingStatus:   []types.ListingStatus{types.ListingActive},
		IncomeRange:     []types.IncomeRange{types.Between25kAnd50k, types.Between50kAnd75k, types.Between75kAnd100k, types.Over100k},
		// Re-enable when Prosper fixes their bug here.
		/*PriorProsperLoansLatePaymentsOneMonthPlus: types.Int32Range{
			Max: types.CreateInt32(0),
		},
		PriorProsperLoansBalanceOutstanding: types.Float64Range{
			Max: types.CreateFloat64(0.0),
		},*/
		InquiriesLast6Months: types.Int32Range{Max: types.CreateInt32(3)},
		DtiWprosperLoan:      types.Float64Range{Max: types.CreateFloat64(0.4)},
		ProsperRating:        []types.ProsperRating{types.RatingAA, types.RatingA, types.RatingB, types.RatingC, types.RatingD, types.RatingE},
	}
	buyer.Poll(1*time.Second, f, *isBuyingEnabled, &c)
	account.Poll(1*time.Minute, &c)
	notes.Poll(10*time.Minute, &c)
	for {
		time.Sleep(10 * time.Minute)
	}
}
