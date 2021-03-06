package notes

import (
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

func Poll(pollInterval time.Duration, nf prosper.NoteFetcher) error {
	log.Printf("starting note polling")
	notes := make(chan prosper.Note)
	notePoller := notePoller{
		nf:           nf,
		notes:        notes,
		pollInterval: pollInterval,
	}
	redisLogger, err := newRedisLogger(notes)
	if err != nil {
		return err
	}

	go redisLogger.Run()
	go notePoller.Run()

	return nil
}
