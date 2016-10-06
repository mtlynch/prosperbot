package buyer

import (
	"reflect"
	"testing"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockRedisSetNXer struct {
	values map[string]string
	err    error
}

func (r *mockRedisSetNXer) SetNX(key string, value interface{}) (bool, error) {
	_, exists := r.values[key]
	if exists {
		return false, r.err
	}
	v, ok := value.(string)
	if !ok {
		panic("unexpected value type")
	}
	r.values[key] = v
	return true, r.err
}

var (
	listingA = prosper.Listing{ListingNumber: 123}
	listingB = prosper.Listing{ListingNumber: 456}
)

func TestSeenListingFilter(t *testing.T) {
	var tests = []struct {
		redisStartingValues map[string]string
		redisErr            error
		listings            []prosper.Listing
		wantNewListings     []prosper.Listing
		msg                 string
	}{
		{
			redisStartingValues: make(map[string]string),
			listings:            []prosper.Listing{listingA},
			wantNewListings:     []prosper.Listing{listingA},
			msg:                 "new listing should pass filter",
		},
		{
			redisStartingValues: make(map[string]string),
			listings:            []prosper.Listing{listingA, listingA, listingA},
			wantNewListings:     []prosper.Listing{listingA},
			msg:                 "repeated instances of same listing should not pass filter",
		},
		{
			redisStartingValues: make(map[string]string),
			listings:            []prosper.Listing{listingA, listingB, listingA},
			wantNewListings:     []prosper.Listing{listingA, listingB},
			msg:                 "repeated instances of same listing should not pass filter",
		},
		{
			redisStartingValues: map[string]string{
				"listing:123": "dummy serialized listing",
				"listing:456": "dummy serialized listing",
			},
			listings:        []prosper.Listing{listingA, listingB},
			wantNewListings: []prosper.Listing{},
			msg:             "previously seen listings should not pass filter",
		},
	}
	for _, tt := range tests {
		listings := make(chan prosper.Listing)
		newListings := make(chan prosper.Listing)
		setNXer := mockRedisSetNXer{
			values: tt.redisStartingValues,
			err:    tt.redisErr,
		}
		filter := seenListingFilter{
			listings:    listings,
			newListings: newListings,
			redis:       &setNXer,
		}
		go filter.Run()
		for _, u := range tt.listings {
			listings <- u
		}
		gotNewListings := []prosper.Listing{}
		for i := 0; i < len(tt.wantNewListings); i++ {
			gotNewListings = append(gotNewListings, <-newListings)
		}
		if !reflect.DeepEqual(gotNewListings, tt.wantNewListings) {
			t.Errorf("%s: unexpected new listings. got = %+v, want = %+v", tt.msg, gotNewListings, tt.wantNewListings)
		}
	}
}
