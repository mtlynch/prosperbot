package notes

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockRedisListPrepender struct {
	LRangeErr   error
	LPushCalled bool
	LPushErr    error
	State       map[string][]string
}

func (p *mockRedisListPrepender) LRange(key string, start int64, stop int64) ([]string, error) {
	if p.LRangeErr != nil {
		return []string{}, p.LRangeErr
	}
	list, ok := p.State[key]
	if !ok {
		return []string{}, nil
	}
	if (start > int64(len(list))) || ((stop + 1) > int64(len(list))) {
		return []string{}, nil
	}
	return list[start : stop+1], nil
}

func (p *mockRedisListPrepender) LPush(key string, values ...interface{}) (int64, error) {
	p.LPushCalled = true
	if p.LPushErr != nil {
		return 0, p.LPushErr
	}
	if _, exists := p.State[key]; !exists {
		p.State[key] = []string{}
	}
	for _, v := range values {
		p.State[key] = append([]string{v.(string)}, p.State[key]...)
	}
	return 0, p.LPushErr
}

type mockClock struct {
	now time.Time
}

func (c mockClock) Now() time.Time {
	return c.now
}

const (
	noteASerializedOld     = `{"Note":{"AgeInMonths":0,"AmountBorrowed":0,"BorrowerRate":0,"DaysPastDue":0,"DebtSaleProceedsReceivedProRataShare":0,"InterestPaidProRataShare":0,"IsSold":false,"LateFeesPaidProRataShare":0,"ListingNumber":0,"LoanNoteID":"noteA","LoanNumber":0,"NextPaymentDueAmountProRataShare":0,"NextPaymentDueDate":"0001-01-01T00:00:00Z","NoteDefaultReasonDescription":"","NoteDefaultReason":null,"NoteOwnershipAmount":0,"NoteSaleFeesPaid":0,"NoteSaleGrossAmountReceived":0,"NoteStatusDescription":"","NoteStatus":0,"OriginationDate":"0001-01-01T00:00:00Z","PrincipalBalanceProRataShare":0,"PrincipalPaidProRataShare":0,"ProsperFeesPaidProRataShare":0,"Rating":0,"ServiceFeesPaidProRataShare":0,"Term":0},"Timestamp":"2016-03-04T23:19:22.000000022Z"}`
	noteASerializedNew     = `{"Note":{"AgeInMonths":0,"AmountBorrowed":0,"BorrowerRate":0,"DaysPastDue":0,"DebtSaleProceedsReceivedProRataShare":0,"InterestPaidProRataShare":0,"IsSold":false,"LateFeesPaidProRataShare":0,"ListingNumber":0,"LoanNoteID":"noteA","LoanNumber":0,"NextPaymentDueAmountProRataShare":0,"NextPaymentDueDate":"0001-01-01T00:00:00Z","NoteDefaultReasonDescription":"","NoteDefaultReason":null,"NoteOwnershipAmount":0,"NoteSaleFeesPaid":0,"NoteSaleGrossAmountReceived":0,"NoteStatusDescription":"","NoteStatus":0,"OriginationDate":"0001-01-01T00:00:00Z","PrincipalBalanceProRataShare":0,"PrincipalPaidProRataShare":0,"ProsperFeesPaidProRataShare":0,"Rating":0,"ServiceFeesPaidProRataShare":0,"Term":0},"Timestamp":"2016-03-05T11:40:15.000000022Z"}`
	noteAChangedSerialized = `{"Note":{"AgeInMonths":1,"AmountBorrowed":0,"BorrowerRate":0,"DaysPastDue":0,"DebtSaleProceedsReceivedProRataShare":0,"InterestPaidProRataShare":0,"IsSold":false,"LateFeesPaidProRataShare":0,"ListingNumber":0,"LoanNoteID":"noteA","LoanNumber":0,"NextPaymentDueAmountProRataShare":0,"NextPaymentDueDate":"0001-01-01T00:00:00Z","NoteDefaultReasonDescription":"","NoteDefaultReason":null,"NoteOwnershipAmount":0,"NoteSaleFeesPaid":0,"NoteSaleGrossAmountReceived":0,"NoteStatusDescription":"","NoteStatus":0,"OriginationDate":"0001-01-01T00:00:00Z","PrincipalBalanceProRataShare":0,"PrincipalPaidProRataShare":0,"ProsperFeesPaidProRataShare":0,"Rating":0,"ServiceFeesPaidProRataShare":0,"Term":0},"Timestamp":"2016-03-05T11:40:15.000000022Z"}`
	noteBSerialized        = `{"Note":{"AgeInMonths":0,"AmountBorrowed":0,"BorrowerRate":0,"DaysPastDue":0,"DebtSaleProceedsReceivedProRataShare":0,"InterestPaidProRataShare":0,"IsSold":false,"LateFeesPaidProRataShare":0,"ListingNumber":0,"LoanNoteID":"noteB","LoanNumber":0,"NextPaymentDueAmountProRataShare":0,"NextPaymentDueDate":"0001-01-01T00:00:00Z","NoteDefaultReasonDescription":"","NoteDefaultReason":2,"NoteOwnershipAmount":0,"NoteSaleFeesPaid":0,"NoteSaleGrossAmountReceived":0,"NoteStatusDescription":"","NoteStatus":0,"OriginationDate":"0001-01-01T00:00:00Z","PrincipalBalanceProRataShare":0,"PrincipalPaidProRataShare":0,"ProsperFeesPaidProRataShare":0,"Rating":0,"ServiceFeesPaidProRataShare":0,"Term":0},"Timestamp":"2016-03-05T11:40:15.000000022Z"}`
	badJSON                = "{{mock bad JSON"
)

var (
	bankruptcy   = prosper.Bankruptcy
	noteA        = prosper.Note{LoanNoteID: "noteA", AgeInMonths: 0}
	noteAChanged = prosper.Note{LoanNoteID: "noteA", AgeInMonths: 1}
	noteB        = prosper.Note{
		LoanNoteID:        "noteB",
		NoteDefaultReason: &bankruptcy,
	}
	redisErr = errors.New("mock redis error")
)

func TestRedisLogger(t *testing.T) {
	var tests = []struct {
		startingRedisState map[string][]string
		updates            []prosper.Note
		lrangeErr          error
		lpushErr           error
		wantLPushCalled    bool
		wantEndState       map[string][]string
		msg                string
	}{
		{
			startingRedisState: map[string][]string{},
			updates:            []prosper.Note{noteA},
			wantLPushCalled:    true,
			wantEndState: map[string][]string{
				"note:noteA": {noteASerializedNew},
			},
			msg: "any update should cause a save when no history exists",
		},
		{
			startingRedisState: map[string][]string{},
			updates:            []prosper.Note{noteA, noteB},
			wantLPushCalled:    true,
			wantEndState: map[string][]string{
				"note:noteA": {noteASerializedNew},
				"note:noteB": {noteBSerialized},
			},
			msg: "any update should cause a save when no history exists and multiple notes are found",
		},
		{
			startingRedisState: map[string][]string{
				"note:noteA": {noteASerializedOld},
			},
			updates:         []prosper.Note{noteA},
			wantLPushCalled: false,
			wantEndState: map[string][]string{
				"note:noteA": {noteASerializedOld},
			},
			msg: "if there is a note update, but no change, don't push",
		},
		{
			startingRedisState: map[string][]string{
				"note:noteB": {noteBSerialized},
			},
			updates:         []prosper.Note{noteB},
			wantLPushCalled: false,
			wantEndState: map[string][]string{
				"note:noteB": {noteBSerialized},
			},
			msg: "if there is a note update, but no change (including on a pointer field), don't push",
		},
		{
			startingRedisState: map[string][]string{
				"note:noteA": {noteASerializedOld},
			},
			updates:         []prosper.Note{noteAChanged},
			wantLPushCalled: true,
			wantEndState: map[string][]string{
				"note:noteA": {noteAChangedSerialized, noteASerializedOld},
			},
			msg: "if there is a note update with changes, push",
		},
		{
			startingRedisState: map[string][]string{},
			updates:            []prosper.Note{noteA},
			lrangeErr:          redisErr,
			wantLPushCalled:    false,
			wantEndState:       map[string][]string{},
			msg:                "when LRange fails, ignore and continue",
		},
		{
			startingRedisState: map[string][]string{},
			updates:            []prosper.Note{noteA},
			lpushErr:           redisErr,
			wantLPushCalled:    true,
			wantEndState:       map[string][]string{},
			msg:                "when LPush fails, ignore and continue",
		},
	}
	for _, tt := range tests {
		noteUpdates := make(chan prosper.Note)
		done := make(chan bool)
		prepender := mockRedisListPrepender{
			State:     tt.startingRedisState,
			LRangeErr: tt.lrangeErr,
			LPushErr:  tt.lpushErr,
		}
		redisLogger := redisLogger{
			noteUpdates: noteUpdates,
			done:        done,
			redis:       &prepender,
			clock:       mockClock{time.Date(2016, 3, 5, 11, 40, 15, 22, time.UTC)},
		}
		go redisLogger.Run()
		for _, u := range tt.updates {
			noteUpdates <- u
		}
		close(noteUpdates)
		<-done
		if prepender.LPushCalled != tt.wantLPushCalled {
			t.Errorf("%s: unexpected LPush call. got: %v, want: %v", tt.msg, prepender.LPushCalled, tt.wantLPushCalled)
		}
		if !reflect.DeepEqual(prepender.State, tt.wantEndState) {
			t.Errorf("%s: unexpected ending state for redis. got: %+v,\t want: %+v", tt.msg, prepender.State, tt.wantEndState)
		}
	}
}

func TestNoteEqual(t *testing.T) {
	tests := []struct {
		a    prosper.Note
		b    prosper.Note
		want bool
	}{
		{prosper.Note{}, prosper.Note{}, true},
		{
			a: prosper.Note{
				AgeInMonths:                          3,
				AmountBorrowed:                       7500,
				BorrowerRate:                         0.3125,
				DaysPastDue:                          0,
				DebtSaleProceedsReceivedProRataShare: 0,
				InterestPaidProRataShare:             0,
				IsSold: false,
				LateFeesPaidProRataShare:         0,
				ListingNumber:                    994439,
				LoanNoteID:                       "1619-2",
				LoanNumber:                       1619,
				NextPaymentDueAmountProRataShare: 308.441467,
				NextPaymentDueDate:               time.Date(2016, 4, 1, 12, 0, 0, 0, time.UTC),
				NoteDefaultReasonDescription:     "",
				NoteDefaultReason:                nil,
				NoteOwnershipAmount:              7150,
				NoteSaleFeesPaid:                 0,
				NoteSaleGrossAmountReceived:      0,
				NoteStatusDescription:            "CURRENT",
				NoteStatus:                       1,
				OriginationDate:                  time.Date(2016, 3, 5, 12, 15, 12, 4, time.UTC),
				PrincipalBalanceProRataShare:     7150,
				PrincipalPaidProRataShare:        0,
				ProsperFeesPaidProRataShare:      0,
				Rating: 6,
				ServiceFeesPaidProRataShare: 0,
				Term: 36,
			},
			b: prosper.Note{
				AgeInMonths:                          3,
				AmountBorrowed:                       7500,
				BorrowerRate:                         0.3125,
				DaysPastDue:                          0,
				DebtSaleProceedsReceivedProRataShare: 0,
				InterestPaidProRataShare:             0,
				IsSold: false,
				LateFeesPaidProRataShare:         0,
				ListingNumber:                    994439,
				LoanNoteID:                       "1619-2",
				LoanNumber:                       1619,
				NextPaymentDueAmountProRataShare: 308.441467,
				NextPaymentDueDate:               time.Date(2016, 4, 1, 12, 0, 0, 0, time.UTC),
				NoteDefaultReasonDescription:     "",
				NoteDefaultReason:                nil,
				NoteOwnershipAmount:              7150,
				NoteSaleFeesPaid:                 0,
				NoteSaleGrossAmountReceived:      0,
				NoteStatusDescription:            "CURRENT",
				NoteStatus:                       1,
				OriginationDate:                  time.Date(2016, 3, 5, 12, 15, 12, 4, time.UTC),
				PrincipalBalanceProRataShare:     7150,
				PrincipalPaidProRataShare:        0,
				ProsperFeesPaidProRataShare:      0,
				Rating: 6,
				ServiceFeesPaidProRataShare: 0,
				Term: 36,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		if got := noteEqual(tt.a, tt.b); got != tt.want {
			t.Errorf("Unexpected equality for notes %+v and %+v. got: %v, want: %v", tt.a, tt.b, got, tt.want)
		}
	}
}
