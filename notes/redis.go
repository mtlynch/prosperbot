package notes

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"

	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/clock"
	"github.com/mtlynch/prosperbot/redis"
)

type redisLogger struct {
	noteUpdates <-chan prosper.Note
	done        chan<- bool
	redis       redis.RedisListPrepender
	clock       clock.Clock
}

func newRedisLogger(noteUpdates <-chan prosper.Note) (redisLogger, error) {
	r, err := redis.New()
	if err != nil {
		return redisLogger{}, err
	}
	return redisLogger{
		noteUpdates: noteUpdates,
		redis:       r,
		clock:       clock.DefaultClock{},
	}, nil
}

var errNotFound = errors.New("note not found")

func noteEqual(a, b prosper.Note) bool {
	if a.AgeInMonths != b.AgeInMonths {
		return false
	}
	if a.AmountBorrowed != b.AmountBorrowed {
		return false
	}
	if a.BorrowerRate != b.BorrowerRate {
		return false
	}
	if a.DaysPastDue != b.DaysPastDue {
		return false
	}
	if a.DebtSaleProceedsReceivedProRataShare != b.DebtSaleProceedsReceivedProRataShare {
		return false
	}
	if a.InterestPaidProRataShare != b.InterestPaidProRataShare {
		return false
	}
	if a.IsSold != b.IsSold {
		return false
	}
	if a.LateFeesPaidProRataShare != b.LateFeesPaidProRataShare {
		return false
	}
	if a.ListingNumber != b.ListingNumber {
		return false
	}
	if a.LoanNoteID != b.LoanNoteID {
		return false
	}
	if a.LoanNumber != b.LoanNumber {
		return false
	}
	if a.NextPaymentDueAmountProRataShare != b.NextPaymentDueAmountProRataShare {
		return false
	}
	if !a.NextPaymentDueDate.Equal(b.NextPaymentDueDate) {
		return false
	}
	if a.NoteDefaultReasonDescription != b.NoteDefaultReasonDescription {
		return false
	}
	if !reflect.DeepEqual(a.NoteDefaultReason, b.NoteDefaultReason) {
		return false
	}
	if a.NoteOwnershipAmount != b.NoteOwnershipAmount {
		return false
	}
	if a.NoteSaleFeesPaid != b.NoteSaleFeesPaid {
		return false
	}
	if a.NoteSaleGrossAmountReceived != b.NoteSaleGrossAmountReceived {
		return false
	}
	if a.NoteStatusDescription != b.NoteStatusDescription {
		return false
	}
	if a.NoteStatus != b.NoteStatus {
		return false
	}
	if !a.OriginationDate.Equal(b.OriginationDate) {
		return false
	}
	if a.PrincipalBalanceProRataShare != b.PrincipalBalanceProRataShare {
		return false
	}
	if a.PrincipalPaidProRataShare != b.PrincipalPaidProRataShare {
		return false
	}
	if a.ProsperFeesPaidProRataShare != b.ProsperFeesPaidProRataShare {
		return false
	}
	if a.Rating != b.Rating {
		return false
	}
	if a.ServiceFeesPaidProRataShare != b.ServiceFeesPaidProRataShare {
		return false
	}
	if a.Term != b.Term {
		return false
	}
	return true
}

func NewRedisLogger(updates <-chan prosper.Note) (redisLogger, error) {
	r, err := redis.New()
	if err != nil {
		return redisLogger{}, err
	}
	done := make(chan bool)
	return redisLogger{
		noteUpdates: updates,
		done:        done,
		redis:       r,
		clock:       clock.DefaultClock{},
	}, nil
}

func (r redisLogger) Run() {
	for {
		n, more := <-r.noteUpdates
		if !more {
			r.done <- true
			return
		}
		nSaved, err := r.getLatestNoteState(n)
		if err == errNotFound {
			// New note, proceed.
		} else if err != nil {
			log.Printf("failed to get note history for %v, err: %v", n.LoanNoteID, err)
			continue
		} else {
			if noteEqual(n, nSaved) {
				// No change, don't save.
				continue
			}
		}
		log.Printf("update to note: %v", n.LoanNoteID)
		if err = r.saveNoteState(n); err != nil {
			log.Printf("failed to save note %+v, err: %v", n, err)
		}
	}
}

func noteToKey(n prosper.Note) string {
	return redis.KeyPrefixNote + n.LoanNoteID
}

func (r redisLogger) getLatestNoteState(n prosper.Note) (prosper.Note, error) {
	noteSerialized, err := r.redis.LRange(noteToKey(n), 0, 0)
	if err != nil {
		return prosper.Note{}, err
	}
	if len(noteSerialized) < 1 {
		return prosper.Note{}, errNotFound
	}
	var record redis.NoteRecord
	err = json.Unmarshal([]byte(noteSerialized[0]), &record)
	if err != nil {
		return prosper.Note{}, err
	}
	return record.Note, nil
}

func (r redisLogger) saveNoteState(n prosper.Note) error {
	record := redis.NoteRecord{
		Note:      n,
		Timestamp: r.clock.Now(),
	}
	serialized, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = r.redis.LPush(noteToKey(n), string(serialized))
	if err != nil {
		return err
	}
	return nil
}
