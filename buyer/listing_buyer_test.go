package buyer

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/mtlynch/gofn-prosper/types"
)

type mockBidPlacer struct {
	gotListingID types.ListingNumber
	gotBidAmount float64
	orderIDs     types.OrderIDs
	errs         []error
}

func (bp *mockBidPlacer) PlaceBid(listingID types.ListingNumber, bidAmount float64) (types.OrderResponse, error) {
	bp.gotListingID = listingID
	bp.gotBidAmount = bidAmount
	var orderID types.OrderID
	orderID, bp.orderIDs = bp.orderIDs[0], bp.orderIDs[1:]
	var err error
	err, bp.errs = bp.errs[0], bp.errs[1:]
	return types.OrderResponse{OrderID: orderID}, err
}

var (
	listingIDA = types.ListingNumber(123)
	listingIDB = types.ListingNumber(456)
	orderIDA   = types.OrderID("order-a")
	orderIDB   = types.OrderID("order-b")
	genericErr = errors.New("generic mock error")
)

func TestListingBuyer(t *testing.T) {
	var tests = []struct {
		listings        []types.Listing
		emittedOrderIDs types.OrderIDs
		emittedErrs     []error
		wantOrderIDs    types.OrderIDs
		msg             string
	}{
		{
			listings: []types.Listing{
				{ListingNumber: listingIDA},
			},
			emittedOrderIDs: types.OrderIDs{orderIDA},
			emittedErrs:     []error{nil},
			wantOrderIDs:    types.OrderIDs{orderIDA},
			msg:             "single listing should result in single order ID",
		},
		{
			listings: []types.Listing{
				{ListingNumber: listingIDA},
				{ListingNumber: listingIDB},
			},
			emittedOrderIDs: types.OrderIDs{orderIDA, orderIDB},
			emittedErrs:     []error{nil, nil},
			wantOrderIDs:    types.OrderIDs{orderIDA, orderIDB},
			msg:             "two listings should result in two order IDs",
		},
		{
			listings: []types.Listing{
				{ListingNumber: listingIDA},
				{ListingNumber: listingIDB},
			},
			emittedOrderIDs: types.OrderIDs{orderIDA, orderIDB},
			emittedErrs:     []error{genericErr, nil},
			wantOrderIDs:    types.OrderIDs{orderIDB},
			msg:             "failed orders should not be reported",
		},
	}
	for _, tt := range tests {
		listings := make(chan types.Listing)
		orderIDs := make(chan types.OrderID)
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
		gotOrderIDs := types.OrderIDs{}
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
