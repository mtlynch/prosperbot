package buyer

import (
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/types"
)

type mockOrderStatusQuerier struct {
	gotOrderID    types.OrderID
	orderStatuses []types.OrderResponse
	errs          []error
}

func (bp *mockOrderStatusQuerier) OrderStatus(orderID types.OrderID) (types.OrderResponse, error) {
	bp.gotOrderID = orderID
	var orderStatus types.OrderResponse
	orderStatus, bp.orderStatuses = bp.orderStatuses[0], bp.orderStatuses[1:]
	var err error
	err, bp.errs = bp.errs[0], bp.errs[1:]
	// Hack to get order statuses to come back in the expected order. Otherwise,
	// the last loop iteration's gofunc in the order status worker tends to spawn
	// first and screw up the ordering of the expected status responses.
	time.Sleep(1 * time.Millisecond)
	return orderStatus, err
}

var (
	orderStatusA = types.OrderResponse{
		OrderStatus: types.OrderInProgress,
		BidStatus:   []types.BidStatus{{Result: types.NoBidResult}},
	}
	orderStatusB = types.OrderResponse{
		OrderStatus: types.OrderCompleted,
		BidStatus:   []types.BidStatus{{Result: types.BidFailed}},
	}
	orderStatusC = types.OrderResponse{
		OrderStatus: types.OrderInProgress,
		BidStatus:   []types.BidStatus{{Result: types.BidSucceeded}},
	}
)

func TestOrderStatusQueryWorker(t *testing.T) {
	var tests = []struct {
		orderID              types.OrderID
		emittedOrderStatuses []types.OrderResponse
		emittedErrs          []error
		wantOrderStatuses    []types.OrderResponse
		msg                  string
	}{
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []types.OrderResponse{orderStatusB},
			emittedErrs:          []error{nil},
			wantOrderStatuses:    []types.OrderResponse{orderStatusB},
			msg:                  "if first status is completed, we're done immediately",
		},
		{
			orderID:              orderIDB,
			emittedOrderStatuses: []types.OrderResponse{orderStatusB},
			emittedErrs:          []error{nil},
			wantOrderStatuses:    []types.OrderResponse{orderStatusB},
			msg:                  "verify we're passing along the correct order ID",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []types.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{nil, nil, nil},
			wantOrderStatuses:    []types.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			msg:                  "query until we get a completed status",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []types.OrderResponse{orderStatusA, orderStatusA, orderStatusC},
			emittedErrs:          []error{nil, nil, nil},
			wantOrderStatuses:    []types.OrderResponse{orderStatusA, orderStatusA, orderStatusC},
			msg:                  "query until we get a completed bid result",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []types.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{genericErr, genericErr, genericErr},
			wantOrderStatuses:    []types.OrderResponse{},
			msg:                  "don't pass along error responses and error out after three",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []types.OrderResponse{orderStatusA, orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{genericErr, genericErr, nil, nil},
			wantOrderStatuses:    []types.OrderResponse{orderStatusA, orderStatusB},
			msg:                  "if there are only two errors, recover",
		},
	}
	for _, tt := range tests {
		orderQuerier := mockOrderStatusQuerier{
			orderStatuses: tt.emittedOrderStatuses,
			errs:          tt.emittedErrs,
		}
		orderStatuses := make(chan types.OrderResponse)
		queryWorker := orderStatusQueryWorker{
			querier:      &orderQuerier,
			orderUpdates: orderStatuses,
		}
		queryWorker.QueryUntilComplete(tt.orderID)
		gotOrderStatuses := []types.OrderResponse{}
		for i := 0; i < len(tt.wantOrderStatuses); i++ {
			gotOrderStatuses = append(gotOrderStatuses, <-orderStatuses)
		}
		if orderQuerier.gotOrderID != tt.orderID {
			t.Errorf("%s: unexpected order ID. got = %+v, want = %+v", tt.msg, orderQuerier.gotOrderID, tt.orderID)
		}
		if !reflect.DeepEqual(gotOrderStatuses, tt.wantOrderStatuses) {
			t.Errorf("%s: unexpected new listings. got = %+v, want = %+v", tt.msg, gotOrderStatuses, tt.wantOrderStatuses)
		}
	}
}
