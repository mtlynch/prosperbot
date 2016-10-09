package buyer

import (
	"reflect"
	"sort"
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

// byListingNumber implements sort.Interface for []propser.Listing based on the
// ListingNumber field.
type byListingNumber []prosper.Listing

func (b byListingNumber) Len() int           { return len(b) }
func (b byListingNumber) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byListingNumber) Less(i, j int) bool { return b[i].ListingNumber < b[j].ListingNumber }

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
			msg:                 "repeated, consecutive instances of same listing should not pass filter",
		},
		{
			redisStartingValues: make(map[string]string),
			listings:            []prosper.Listing{listingA, listingB, listingA},
			wantNewListings:     []prosper.Listing{listingA, listingB},
			msg:                 "repeated, nonconsecutive instances of same listing should not pass filter",
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
		go func() {
			for _, u := range tt.listings {
				listings <- u
			}
			close(listings)
		}()
		filter.Run()
		gotNewListings := []prosper.Listing{}
		for i := 0; i < len(tt.wantNewListings); i++ {
			gotNewListings = append(gotNewListings, <-newListings)
		}
		sort.Sort(byListingNumber(gotNewListings))
		if !reflect.DeepEqual(gotNewListings, tt.wantNewListings) {
			t.Errorf("%s: unexpected new listings. got = %+v, want = %+v", tt.msg, gotNewListings, tt.wantNewListings)
		}
	}
}
