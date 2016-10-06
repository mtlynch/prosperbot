package redis

import (
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type (
	AccountRecord struct {
		Value     prosper.AccountInformation
		Timestamp time.Time
	}
	NoteRecord struct {
		Note      prosper.Note
		Timestamp time.Time
	}
	OrderRecord struct {
		Order     prosper.OrderResponse
		Timestamp time.Time
	}
)
