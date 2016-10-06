package buyer

import (
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockOrderStatusQuerier struct {
	gotOrderID    prosper.OrderID
	orderStatuses []prosper.OrderResponse
	errs          []error
}

func (bp *mockOrderStatusQuerier) OrderStatus(orderID prosper.OrderID) (prosper.OrderResponse, error) {
	bp.gotOrderID = orderID
	var orderStatus prosper.OrderResponse
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
	orderStatusA = prosper.OrderResponse{
		OrderStatus: prosper.OrderInProgress,
		BidStatus:   []prosper.BidStatus{{Result: prosper.NoBidResult}},
	}
	orderStatusB = prosper.OrderResponse{
		OrderStatus: prosper.OrderCompleted,
		BidStatus:   []prosper.BidStatus{{Result: prosper.BidFailed}},
	}
	orderStatusC = prosper.OrderResponse{
		OrderStatus: prosper.OrderInProgress,
		BidStatus:   []prosper.BidStatus{{Result: prosper.BidSucceeded}},
	}
)

func TestOrderStatusQueryWorker(t *testing.T) {
	var tests = []struct {
		orderID              prosper.OrderID
		emittedOrderStatuses []prosper.OrderResponse
		emittedErrs          []error
		wantOrderStatuses    []prosper.OrderResponse
		msg                  string
	}{
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusB},
			emittedErrs:          []error{nil},
			wantOrderStatuses:    []prosper.OrderResponse{orderStatusB},
			msg:                  "if first status is completed, we're done immediately",
		},
		{
			orderID:              orderIDB,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusB},
			emittedErrs:          []error{nil},
			wantOrderStatuses:    []prosper.OrderResponse{orderStatusB},
			msg:                  "verify we're passing along the correct order ID",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{nil, nil, nil},
			wantOrderStatuses:    []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			msg:                  "query until we get a completed status",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusC},
			emittedErrs:          []error{nil, nil, nil},
			wantOrderStatuses:    []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusC},
			msg:                  "query until we get a completed bid result",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{genericErr, genericErr, genericErr},
			wantOrderStatuses:    []prosper.OrderResponse{},
			msg:                  "don't pass along error responses and error out after three",
		},
		{
			orderID:              orderIDA,
			emittedOrderStatuses: []prosper.OrderResponse{orderStatusA, orderStatusA, orderStatusA, orderStatusB},
			emittedErrs:          []error{genericErr, genericErr, nil, nil},
			wantOrderStatuses:    []prosper.OrderResponse{orderStatusA, orderStatusB},
			msg:                  "if there are only two errors, recover",
		},
	}
	for _, tt := range tests {
		orderQuerier := mockOrderStatusQuerier{
			orderStatuses: tt.emittedOrderStatuses,
			errs:          tt.emittedErrs,
		}
		orderStatuses := make(chan prosper.OrderResponse)
		queryWorker := orderStatusQueryWorker{
			querier:      &orderQuerier,
			orderUpdates: orderStatuses,
		}
		queryWorker.QueryUntilComplete(tt.orderID)
		gotOrderStatuses := []prosper.OrderResponse{}
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
