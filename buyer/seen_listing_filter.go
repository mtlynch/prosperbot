package buyer

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/redis"
)

type seenListingFilter struct {
	listings    <-chan prosper.Listing
	newListings chan<- prosper.Listing
	redis       redis.RedisSetNXer
}

func NewSeenListingFilter(listings <-chan prosper.Listing, newListings chan<- prosper.Listing) (seenListingFilter, error) {
	r, err := redis.New()
	if err != nil {
		return seenListingFilter{}, err
	}
	return seenListingFilter{
		listings:    listings,
		newListings: newListings,
		redis:       r,
	}, nil
}

func (r seenListingFilter) Run() {
	for {
		listing := <-r.listings
		isNew, err := r.saveListing(listing)
		if err != nil {
			log.Printf("failed to save listing: %v", err)
			continue
		}
		if !isNew {
			continue
		}
		log.Printf("found new listing: %v", listing.ListingNumber)
		go func() { r.newListings <- listing }()
	}
}

func (r seenListingFilter) saveListing(listing prosper.Listing) (isNew bool, err error) {
	serialized, err := json.Marshal(listing)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf("%s%d", redis.KeyPrefixListing, listing.ListingNumber)
	return r.redis.SetNX(key, string(serialized))
}
