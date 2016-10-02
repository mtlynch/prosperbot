package buyer

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/types"
)

type mockRedisSetter struct {
	Values  map[string]string
	SetErrs []error
}

func (m *mockRedisSetter) Set(key string, value interface{}) (string, error) {
	var err error
	err, m.SetErrs = m.SetErrs[0], m.SetErrs[1:]
	if err == nil {
		m.Values[key] = value.(string)
	}
	return "", err
}

const (
	orderAUpdate1Serialized = `{"Order":{"OrderID":"id-a","BidStatus":[{"ListingID":54321,"BidAmount":25,"Status":0,"Result":0,"BidAmountPlaced":25}],"OrderStatus":0,"OrderDate":"2016-04-23T11:54:29Z"},"Timestamp":"2016-02-14T12:28:15.000000022Z"}`
	orderAUpdate2Serialized = `{"Order":{"OrderID":"id-a","BidStatus":[{"ListingID":54321,"BidAmount":25,"Status":0,"Result":4,"BidAmountPlaced":25}],"OrderStatus":0,"OrderDate":"2016-04-23T11:54:29Z"},"Timestamp":"2016-02-14T12:28:15.000000022Z"}`
	orderBSerialized        = `{"Order":{"OrderID":"id-b","BidStatus":[{"ListingID":987654,"BidAmount":37.5,"Status":0,"Result":3,"BidAmountPlaced":37.5}],"OrderStatus":0,"OrderDate":"2016-03-25T20:18:04.000000036Z"},"Timestamp":"2016-02-14T12:28:15.000000022Z"}`
)

var (
	mockErr       = errors.New("mock error")
	orderAUpdate1 = types.OrderResponse{
		OrderID: "id-a",
		BidStatus: []types.BidStatus{
			types.BidStatus{
				BidRequest: types.BidRequest{
					ListingID: types.ListingNumber(54321),
					BidAmount: 25.0,
				},
				Status:          types.Pending,
				Result:          types.NoBidResult,
				BidAmountPlaced: 25.0,
			},
		},
		OrderStatus: types.OrderInProgress,
		OrderDate:   time.Date(2016, 4, 23, 11, 54, 29, 0, time.UTC),
	}
	orderAUpdate2 = types.OrderResponse{
		OrderID: "id-a",
		BidStatus: []types.BidStatus{
			types.BidStatus{
				BidRequest: types.BidRequest{
					ListingID: types.ListingNumber(54321),
					BidAmount: 25.0,
				},
				Status:          types.Pending,
				Result:          types.BidSucceeded,
				BidAmountPlaced: 25.0,
			},
		},
		OrderStatus: types.OrderInProgress,
		OrderDate:   time.Date(2016, 4, 23, 11, 54, 29, 0, time.UTC),
	}
	orderB = types.OrderResponse{
		OrderID: "id-b",
		BidStatus: []types.BidStatus{
			types.BidStatus{
				BidRequest: types.BidRequest{
					ListingID: types.ListingNumber(987654),
					BidAmount: 37.50,
				},
				Status:          types.Pending,
				Result:          types.BidFailed,
				BidAmountPlaced: 37.50,
			},
		},
		OrderStatus: types.OrderInProgress,
		OrderDate:   time.Date(2016, 3, 25, 20, 18, 4, 36, time.UTC),
	}
)

func TestRedisLogger(t *testing.T) {
	var tests = []struct {
		updates    []types.OrderResponse
		setErrs    []error
		wantValues map[string]string
		msg        string
	}{
		{
			updates: []types.OrderResponse{orderAUpdate1},
			setErrs: []error{nil},
			wantValues: map[string]string{
				"order:id-a": orderAUpdate1Serialized,
			},
			msg: "single update should add a single value to redis",
		},
		{
			updates: []types.OrderResponse{orderAUpdate1, orderAUpdate2},
			setErrs: []error{nil, nil},
			wantValues: map[string]string{
				"order:id-a": orderAUpdate2Serialized,
			},
			msg: "multiple updates to same order should add a single value to redis",
		},
		{
			updates: []types.OrderResponse{orderAUpdate1, orderAUpdate2},
			setErrs: []error{mockErr, nil},
			wantValues: map[string]string{
				"order:id-a": orderAUpdate2Serialized,
			},
			msg: "errors setting redis values should be ignored",
		},
		{
			updates: []types.OrderResponse{orderAUpdate1, orderAUpdate2, orderB},
			setErrs: []error{nil, nil, nil},
			wantValues: map[string]string{
				"order:id-a": orderAUpdate2Serialized,
				"order:id-b": orderBSerialized,
			},
			msg: "multiple order updates should succeed",
		},
	}
	for _, tt := range tests {
		orderUpdates := make(chan types.OrderResponse)
		done := make(chan bool)
		mockSetter := mockRedisSetter{
			Values:  map[string]string{},
			SetErrs: tt.setErrs,
		}
		statusLogger := orderStatusLogger{
			redis:        &mockSetter,
			orderUpdates: orderUpdates,
			done:         done,
			clock:        mockClock{time.Date(2016, 2, 14, 12, 28, 15, 22, time.UTC)},
		}
		go statusLogger.Run()
		for _, u := range tt.updates {
			orderUpdates <- u
		}
		close(orderUpdates)
		<-done
		if !reflect.DeepEqual(mockSetter.Values, tt.wantValues) {
			t.Errorf("%s: unexpected values set in redis. got: %v, want: %v", tt.msg, mockSetter.Values, tt.wantValues)
		}
	}
}
