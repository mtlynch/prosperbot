package buyer

import (
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/interval"
	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockListingSearcher struct {
	gotExcludeListingsInvested bool
	gotSearchFilter            prosper.SearchFilter
	calls                      int
	listings                   []prosper.Listing
	err                        error
}

func (ls *mockListingSearcher) Search(sp prosper.SearchParams) (prosper.SearchResponse, error) {
	ls.calls++
	ls.gotExcludeListingsInvested = sp.ExcludeListingsInvested
	ls.gotSearchFilter = sp.Filter
	limit := sp.Offset + sp.Limit
	if limit > len(ls.listings) {
		limit = len(ls.listings)
	}
	results := ls.listings[sp.Offset:limit]
	return prosper.SearchResponse{
		Results:     results,
		ResultCount: len(results),
		TotalCount:  len(ls.listings),
	}, ls.err
}

func makeListings(count int) []prosper.Listing {
	listings := []prosper.Listing{}
	for i := 0; i < count; i++ {
		listings = append(listings, prosper.Listing{ListingNumber: prosper.ListingNumber(i)})
	}
	return listings
}

type mockClock struct {
	now time.Time
}

func (c mockClock) Now() time.Time {
	return c.now
}

var (
	mockCurrentTime   = time.Date(2016, 1, 1, 9, 2, 0, 0, time.UTC)
	mockTimeOneMinAgo = time.Date(2016, 1, 1, 9, 1, 0, 0, time.UTC)
)

func TestListingPoller(t *testing.T) {
	var tests = []struct {
		serverListings []prosper.Listing
		searchFilter   prosper.SearchFilter
		searchErr      error
		wantCalls      int
	}{
		{
			serverListings: makeListings(1),
			searchFilter: prosper.SearchFilter{
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 1,
		},
		{
			serverListings: makeListings(50),
			searchFilter: prosper.SearchFilter{
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 1,
		},
		{
			serverListings: makeListings(51),
			searchFilter: prosper.SearchFilter{
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 2,
		},
		{
			serverListings: makeListings(100),
			searchFilter: prosper.SearchFilter{
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 2,
		},
		{
			serverListings: makeListings(101),
			searchFilter: prosper.SearchFilter{
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 3,
		},
		{
			serverListings: makeListings(1),
			searchFilter: prosper.SearchFilter{
				EstimatedReturn:  interval.NewFloat64Range(0.05, 0.06),
				ListingStartDate: interval.TimeRange{Min: &mockTimeOneMinAgo},
			},
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		listings := make(chan prosper.Listing)
		searcher := mockListingSearcher{
			listings: tt.serverListings,
			err:      tt.searchErr,
		}
		listingPoller := listingPoller{
			s:            &searcher,
			searchFilter: tt.searchFilter,
			listings:     listings,
			pollInterval: 10 * time.Second,
			clock:        mockClock{mockCurrentTime},
		}
		go listingPoller.Run()
		var gotListings []prosper.Listing
		for i := 0; i < len(tt.serverListings); i++ {
			gotListings = append(gotListings, <-listings)
		}
		if !reflect.DeepEqual(tt.serverListings, gotListings) {
			t.Errorf("for listings size %d, unexpected server listings. got: %+v, want: %+v", len(tt.serverListings), gotListings, tt.serverListings)
		}
		if !reflect.DeepEqual(tt.searchFilter, searcher.gotSearchFilter) {
			t.Errorf("for listings size %d, unexpected search filter. got: %+v, want %+v", len(tt.serverListings), searcher.gotSearchFilter, tt.searchFilter)
		}
		if !searcher.gotExcludeListingsInvested {
			t.Errorf("expected listing poller to exclude listings invested")
		}
		if searcher.calls != tt.wantCalls {
			t.Errorf("for listings size %d, unexpected calls to client.Search. got: %d, want: %d", len(tt.serverListings), searcher.calls, tt.wantCalls)
		}
	}
}
