package buyer

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockBidPlacer struct {
	orderIDs prosper.OrderIDs
	errs     []error
}

func (bp *mockBidPlacer) PlaceBid(b prosper.BidRequest) (prosper.OrderResponse, error) {
	var orderID prosper.OrderID
	orderID, bp.orderIDs = bp.orderIDs[0], bp.orderIDs[1:]
	var err error
	err, bp.errs = bp.errs[0], bp.errs[1:]
	return prosper.OrderResponse{OrderID: orderID}, err
}

var (
	listingIDA = prosper.ListingNumber(123)
	listingIDB = prosper.ListingNumber(456)
	orderIDA   = prosper.OrderID("order-a")
	orderIDB   = prosper.OrderID("order-b")
	genericErr = errors.New("generic mock error")
)

func TestListingBuyer(t *testing.T) {
	var tests = []struct {
		listings        []prosper.Listing
		emittedOrderIDs prosper.OrderIDs
		emittedErrs     []error
		wantOrderIDs    prosper.OrderIDs
		msg             string
	}{
		{
			listings: []prosper.Listing{
				{ListingNumber: listingIDA},
			},
			emittedOrderIDs: prosper.OrderIDs{orderIDA},
			emittedErrs:     []error{nil},
			wantOrderIDs:    prosper.OrderIDs{orderIDA},
			msg:             "single listing should result in single order ID",
		},
		{
			listings: []prosper.Listing{
				{ListingNumber: listingIDA},
				{ListingNumber: listingIDB},
			},
			emittedOrderIDs: prosper.OrderIDs{orderIDA, orderIDB},
			emittedErrs:     []error{nil, nil},
			wantOrderIDs:    prosper.OrderIDs{orderIDA, orderIDB},
			msg:             "two listings should result in two order IDs",
		},
		{
			listings: []prosper.Listing{
				{ListingNumber: listingIDA},
				{ListingNumber: listingIDB},
			},
			emittedOrderIDs: prosper.OrderIDs{orderIDA, orderIDB},
			emittedErrs:     []error{genericErr, nil},
			wantOrderIDs:    prosper.OrderIDs{orderIDB},
			msg:             "failed orders should not be reported",
		},
	}
	for _, tt := range tests {
		listings := make(chan prosper.Listing)
		orderIDs := make(chan prosper.OrderID)
		bidPlacer := mockBidPlacer{
			orderIDs: tt.emittedOrderIDs,
			errs:     tt.emittedErrs,
		}
		buyer := listingBuyer{
			listings:  listings,
			orders:    orderIDs,
			bidPlacer: &bidPlacer,
		}
		go func() {
			for _, u := range tt.listings {
				listings <- u
			}
			close(listings)
		}()
		buyer.Run()
		gotOrderIDs := prosper.OrderIDs{}
		for i := 0; i < len(tt.wantOrderIDs); i++ {
			gotOrderIDs = append(gotOrderIDs, <-orderIDs)
		}
		gotOrderIDs.Len()
		sort.Sort(gotOrderIDs)
		if !reflect.DeepEqual(gotOrderIDs, tt.wantOrderIDs) {
			t.Errorf("%s: unexpected new listings. got = %+v, want = %+v", tt.msg, gotOrderIDs, tt.wantOrderIDs)
		}
	}
}
