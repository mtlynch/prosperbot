package buyer

import (
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/clock"
)

// TODO: Add support in Polling for excluding based on a blacklist of
// employment status. Do I actually need to do this? Maybe I can just
// whitelist employment statuses.

func Poll(checkInterval time.Duration, f prosper.SearchFilter, isBuyingEnabled bool, c *prosper.Client) error {
	allListings := make(chan prosper.Listing)
	newListings := make(chan prosper.Listing)
	orders := make(chan prosper.OrderID)
	orderUpdates := make(chan prosper.OrderResponse)
	listingPoller := listingPoller{
		s:            c,
		searchFilter: f,
		listings:     allListings,
		pollInterval: checkInterval,
		clock:        clock.DefaultClock{},
	}
	seenFilter, err := NewSeenListingFilter(allListings, newListings)
	if err != nil {
		return err
	}

	var buyer listingBuyer
	var tracker orderTracker
	var logger orderStatusLogger
	if isBuyingEnabled {
		buyer = listingBuyer{
			listings:  newListings,
			orders:    orders,
			bidPlacer: c,
			bidAmount: 25.0,
		}
		tracker = orderTracker{
			querier:      c,
			orders:       orders,
			orderUpdates: orderUpdates,
		}
		logger, err = NewOrderStatusLogger(orderUpdates)
		if err != nil {
			log.Printf("failed to create order status logger: %v", err)
			return err
		}
	}
	go func() {
		log.Printf("starting buyer polling")

		go listingPoller.Run()
		go seenFilter.Run()
		if isBuyingEnabled {
			go buyer.Run()
			go tracker.Run()
			go logger.Run()
		} else {
			l := <-newListings
			log.Printf("new purchase candidate: %v", l.ListingNumber)
		}
	}()

	return nil
}
