package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/interval"
	"github.com/mtlynch/gofn-prosper/prosper"
	"github.com/mtlynch/gofn-prosper/prosper/auth"

	"github.com/mtlynch/prosperbot/account"
	"github.com/mtlynch/prosperbot/buyer"
	"github.com/mtlynch/prosperbot/notes"
)

func parseCredentials(path string) (creds auth.ClientCredentials, err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return auth.ClientCredentials{}, err
	}
	err = json.Unmarshal(file, &creds)
	if err != nil {
		return auth.ClientCredentials{}, err
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
		EstimatedReturn: interval.Float64Range{Min: interval.CreateFloat64(0.0849)},
		ListingStatus:   []prosper.ListingStatus{prosper.ListingActive},
		IncomeRange:     []prosper.IncomeRange{prosper.Between25kAnd50k, prosper.Between50kAnd75k, prosper.Between75kAnd100k, prosper.Over100k},
		// Re-enable when Prosper fixes their bug here.
		/*PriorProsperLoansLatePaymentsOneMonthPlus: interval.Int32Range{
			Max: interval.CreateInt32(0),
		},
		PriorProsperLoansBalanceOutstanding: interval.Float64Range{
			Max: interval.CreateFloat64(0.0),
		},*/
		InquiriesLast6Months: interval.Int32Range{Max: interval.CreateInt32(3)},
		DtiWprosperLoan:      interval.Float64Range{Max: interval.CreateFloat64(0.4)},
		Rating:               []prosper.Rating{prosper.RatingAA, prosper.RatingA, prosper.RatingB, prosper.RatingC, prosper.RatingD, prosper.RatingE},
	}
	buyer.Poll(1*time.Second, f, *isBuyingEnabled, c)
	account.Poll(1*time.Minute, c)
	notes.Poll(10*time.Minute, c)
	for {
		time.Sleep(10 * time.Minute)
	}
}
