package buyer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mtlynch/gofn-prosper/types"
)

type mockBidPlacer struct {
	gotListingID types.ListingNumber
	gotBidAmount float64
	orderIDs     []types.OrderID
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
		emittedOrderIDs []types.OrderID
		emittedErrs     []error
		wantOrderIDs    []types.OrderID
	}{
		{
			listings: []types.Listing{
				types.Listing{ListingNumber: listingIDA},
			},
			emittedOrderIDs: []types.OrderID{orderIDA},
			emittedErrs:     []error{nil},
			wantOrderIDs:    []types.OrderID{orderIDA},
		},
		{
			listings: []types.Listing{
				types.Listing{ListingNumber: listingIDA},
				types.Listing{ListingNumber: listingIDB},
			},
			emittedOrderIDs: []types.OrderID{orderIDA, orderIDB},
			emittedErrs:     []error{nil, nil},
			wantOrderIDs:    []types.OrderID{orderIDA, orderIDB},
		},
		{
			listings: []types.Listing{
				types.Listing{ListingNumber: listingIDA},
				types.Listing{ListingNumber: listingIDB},
			},
			emittedOrderIDs: []types.OrderID{orderIDA, orderIDB},
			emittedErrs:     []error{genericErr, nil},
			wantOrderIDs:    []types.OrderID{orderIDB},
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
		go buyer.Run()
		for _, u := range tt.listings {
			listings <- u
		}
		gotOrderIDs := []types.OrderID{}
		for i := 0; i < len(tt.wantOrderIDs); i++ {
			gotOrderIDs = append(gotOrderIDs, <-orderIDs)
		}
		if !reflect.DeepEqual(gotOrderIDs, tt.wantOrderIDs) {
			t.Errorf("unexpected new listings. got = %+v, want = %+v", gotOrderIDs, tt.wantOrderIDs)
		}
	}
}
