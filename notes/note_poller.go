package notes

import (
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
	"github.com/mtlynch/gofn-prosper/types"
)

type notePoller struct {
	nf           prosper.NoteFetcher
	notes        chan<- types.Note
	pollInterval time.Duration
}

const MaxAttempts = 3

func (np notePoller) Run() {
	getListings := func() {
		attempts := 0
		offset := 0
		limit := 25
		for {
			if attempts >= MaxAttempts {
				log.Printf("too many note poll failures, bailing out")
				return
			}
			attempts++
			response, err := np.nf.Notes(offset, limit)
			if err != nil {
				log.Printf("failed to get new notes: %v", err)
				continue
			}
			for _, note := range response.Result {
				go func(n types.Note) {
					np.notes <- n
				}(note)
			}
			if int(response.ResultCount) < limit {
				return
			}
			offset += response.ResultCount
			if offset >= response.TotalCount {
				return
			}
			attempts = 0
		}
	}
	go func() {
		for {
			go getListings()
			time.Sleep(np.pollInterval)
		}
	}()
}
