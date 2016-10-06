package buyer

import (
	"log"

	"github.com/mtlynch/gofn-prosper/prosper"
	"github.com/mtlynch/gofn-prosper/types"
)

type listingBuyer struct {
	listings  <-chan prosper.Listing
	orders    chan<- prosper.OrderID
	bidPlacer prosper.BidPlacer
	bidAmount float64
}

func (lb listingBuyer) Run() {
	csf := ClientSideFilter{
		PriorProsperLoansLatePaymentsOneMonthPlus: types.Int32Range{
			Max: types.CreateInt32(0),
		},
		PriorProsperLoansBalanceOutstanding: types.Float64Range{
			Max: types.CreateFloat64(0.0),
		},
		CurrentDelinquencies: types.Int32Range{
			Max: types.CreateInt32(0),
		},
		InquiriesLast6Months: types.Int32Range{
			Max: types.CreateInt32(3),
		},
		EmploymentStatusDescriptionBlacklist: []string{
			"Unemployed", "Not Available",
		},
	}
	for {
		listing, more := <-lb.listings
		if !more {
			return
		}
		// TODO: Do purchase filtering in a cleaner place
		if !csf.Filter(listing) {
			continue
		}

		// TODO: Add in retries.
		orderResponse, err := lb.bidPlacer.PlaceBid(listing.ListingNumber, lb.bidAmount)
		if err != nil {
			log.Printf("failed to place bid on listing %v: %v", listing.ListingNumber, err)
			continue
		}
		log.Printf("placed bid, order ID: %v, listing: %v", orderResponse.OrderID, listing.ListingNumber)
		go func() { lb.orders <- orderResponse.OrderID }()
	}
}
