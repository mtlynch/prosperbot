package account

import (
	"log"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

func Poll(updateInterval time.Duration, accounter prosper.Accounter) error {
	log.Printf("starting account polling")
	accountUpdates := make(chan prosper.AccountInformation)
	logger, err := NewRedisLogger(accountUpdates)
	if err != nil {
		return err
	}
	go logger.Run()
	go func() {
		for {
			a, err := accounter.Account(prosper.AccountParams{})
			if err != nil {
				log.Printf("failed to query account information: %v", err)
			} else {
				accountUpdates <- a
			}
			time.Sleep(updateInterval)
		}
	}()

	return nil
}
