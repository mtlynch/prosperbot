package buyer

import (
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/interval"
	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/clock"
)

type listingPoller struct {
	s            prosper.ListingSearcher
	searchFilter prosper.SearchFilter
	listings     chan<- prosper.Listing
	pollInterval time.Duration
	clock        clock.Clock
}

// Maximum number of attempts before we give up on the listing poll attempt.
const MaxAttempts = 1

func (lp listingPoller) Run() {
	getListings := func() {
		attempts := 0
		offset := 0
		limit := 50
		excludeListingsInvested := true
		timeCutoff := lp.clock.Now().UTC().Add(-1 * time.Minute)
		filter := lp.searchFilter
		filter.ListingStartDate = interval.TimeRange{Min: &timeCutoff}

		for {
			if attempts >= MaxAttempts {
				log.Printf("too many listing poll errors, bailing out")
				return
			}
			attempts++
			response, err := lp.s.Search(prosper.SearchParams{
				Offset: offset,
				Limit:  limit,
				ExcludeListingsInvested: excludeListingsInvested,
				Filter:                  filter,
			})
			if err != nil {
				log.Printf("failed to get new listings: %v", err)
				continue
			}
			for _, listing := range response.Results {
				lp.listings <- listing
			}
			if int(response.ResultCount) < limit {
				return
			}
			offset += response.ResultCount
			if offset >= response.TotalCount {
				return
			}
			attempts = 0
		}
	}
	for {
		go getListings()
		time.Sleep(lp.pollInterval)
	}
}
