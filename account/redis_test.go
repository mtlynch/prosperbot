package account

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/redis"
)

type mockRedisListPrepender struct {
	LRangeKey   string
	LRangeErr   error
	LPushKey    string
	LPushValues []string
	LPushErr    error
	List        []string
}

func (p *mockRedisListPrepender) LRange(key string, start int64, stop int64) ([]string, error) {
	p.LRangeKey = key
	if p.LRangeErr != nil {
		return []string{}, p.LRangeErr
	}
	if (start > int64(len(p.List))) || ((stop + 1) > int64(len(p.List))) {
		return []string{}, nil
	}
	return p.List[start : stop+1], nil
}

func (p *mockRedisListPrepender) LPush(key string, values ...interface{}) (int64, error) {
	p.LPushKey = key
	for _, v := range values {
		p.List = append([]string{v.(string)}, p.List...)
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
	updateASerializedOld = `{"Value":{"AvailableCashBalance":100,"TotalPrincipalReceivedOnActiveNotes":0,"OutstandingPrincipalOnActiveNotes":0,"LastWithdrawAmount":0,"LastDepositAmount":0,"LastDepositDate":"0001-01-01T00:00:00Z","PendingInvestmentsPrimaryMarket":0,"PendingInvestmentsSecondaryMarket":0,"PendingQuickInvestOrders":0,"TotalAmountInvestedOnActiveNotes":0,"TotalAccountValue":0,"InflightGross":0,"LastWithdrawDate":"0001-01-01T00:00:00Z"},"Timestamp":"2016-01-28T15:35:04.000000022Z"}`
	updateASerializedNew = `{"Value":{"AvailableCashBalance":100,"TotalPrincipalReceivedOnActiveNotes":0,"OutstandingPrincipalOnActiveNotes":0,"LastWithdrawAmount":0,"LastDepositAmount":0,"LastDepositDate":"0001-01-01T00:00:00Z","PendingInvestmentsPrimaryMarket":0,"PendingInvestmentsSecondaryMarket":0,"PendingQuickInvestOrders":0,"TotalAmountInvestedOnActiveNotes":0,"TotalAccountValue":0,"InflightGross":0,"LastWithdrawDate":"0001-01-01T00:00:00Z"},"Timestamp":"2016-02-14T12:28:15.000000022Z"}`
	updateBSerializedNew = `{"Value":{"AvailableCashBalance":125.5,"TotalPrincipalReceivedOnActiveNotes":0,"OutstandingPrincipalOnActiveNotes":0,"LastWithdrawAmount":0,"LastDepositAmount":0,"LastDepositDate":"0001-01-01T00:00:00Z","PendingInvestmentsPrimaryMarket":0,"PendingInvestmentsSecondaryMarket":0,"PendingQuickInvestOrders":0,"TotalAmountInvestedOnActiveNotes":0,"TotalAccountValue":0,"InflightGross":0,"LastWithdrawDate":"0001-01-01T00:00:00Z"},"Timestamp":"2016-02-14T12:28:15.000000022Z"}`
	badJSON              = "{{mock bad JSON"
)

var (
	updateA = prosper.AccountInformation{AvailableCashBalance: 100.0}
	updateB = prosper.AccountInformation{AvailableCashBalance: 125.5}
)

func TestRedisLogger(t *testing.T) {
	var tests = []struct {
		startingList    []string
		updates         []prosper.AccountInformation
		lpushErr        error
		wantLPushCalled bool
		wantList        []string
		msg             string
	}{
		{
			updates:         []prosper.AccountInformation{updateA},
			wantLPushCalled: true,
			wantList:        []string{updateASerializedNew},
			msg:             "any update should cause a save when no history exists",
		},
		{
			startingList:    []string{badJSON},
			updates:         []prosper.AccountInformation{updateA},
			wantLPushCalled: true,
			wantList:        []string{updateASerializedNew, badJSON},
			msg:             "error on retrieving latest info should be treated as empty history",
		},
		{
			startingList: []string{updateASerializedOld},
			updates:      []prosper.AccountInformation{updateA},
			wantList:     []string{updateASerializedOld},
			msg:          "an update that is identical to latest data should not cause a save",
		},
		{
			startingList:    []string{updateASerializedOld},
			updates:         []prosper.AccountInformation{updateB},
			wantLPushCalled: true,
			wantList:        []string{updateBSerializedNew, updateASerializedOld},
			msg:             "an update that differs from to latest data should cause a save",
		},
		{
			updates:         []prosper.AccountInformation{updateA, updateA, updateB},
			wantLPushCalled: true,
			wantList:        []string{updateBSerializedNew, updateASerializedNew},
			msg:             "an update of A, A, B should result in saves of A and B",
		},
		{
			updates:         []prosper.AccountInformation{updateA, updateA, updateB},
			lpushErr:        errors.New("mock LPush error"),
			wantLPushCalled: true,
			wantList:        []string{updateBSerializedNew, updateASerializedNew, updateASerializedNew},
			msg:             "when LPush fails, log the error, but continue on",
		},
	}
	for _, tt := range tests {
		accountUpdates := make(chan prosper.AccountInformation)
		prepender := mockRedisListPrepender{
			List:     tt.startingList,
			LPushErr: tt.lpushErr,
		}
		redisLogger := redisLogger{
			accountUpdates: accountUpdates,
			redis:          &prepender,
			clock:          mockClock{time.Date(2016, 2, 14, 12, 28, 15, 22, time.UTC)},
		}

		go func() {
			for _, u := range tt.updates {
				accountUpdates <- u
			}
			close(accountUpdates)
		}()
		redisLogger.Run()
		if prepender.LRangeKey != redis.KeyAccountInformation {
			t.Errorf("%s: unexpected key for LRange. got: %v, want: %v", tt.msg, prepender.LRangeKey, redis.KeyAccountInformation)
		}
		if tt.wantLPushCalled && prepender.LPushKey != redis.KeyAccountInformation {
			t.Errorf("%s: unexpected key for LPush. got: %v, want: %v", tt.msg, prepender.LPushKey, redis.KeyAccountInformation)
		} else if !tt.wantLPushCalled && prepender.LPushKey != "" {
			t.Errorf("%s: unexpected key for LPush. got: %v, want: %v", tt.msg, prepender.LPushKey, nil)
		}
		if !reflect.DeepEqual(prepender.List, tt.wantList) {
			if len(prepender.List) != len(tt.wantList) {
				t.Errorf("%s: unexpected saved value count. got: %d, want: %d", tt.msg, len(prepender.List), len(tt.wantList))
			}
			t.Errorf("%s: unexpected values saved to redis. got = %+v, want = %+v", tt.msg, prepender.List, tt.wantList)
		}
	}
}
