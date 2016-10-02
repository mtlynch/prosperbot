package redis

import (
	"time"

	"github.com/mtlynch/gofn-prosper/types"
)

type (
	AccountRecord struct {
		Value     types.AccountInformation
		Timestamp time.Time
	}
	NoteRecord struct {
		Note      types.Note
		Timestamp time.Time
	}
	OrderRecord struct {
		Order     types.OrderResponse
		Timestamp time.Time
	}
)
