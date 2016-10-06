package account

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/mtlynch/gofn-prosper/prosper"
	"github.com/mtlynch/gofn-prosper/types"

	"github.com/mtlynch/prosperbot/redis"
)

type redisLogger struct {
	accountUpdates <-chan prosper.AccountInformation
	redis          redis.RedisListPrepender
	clock          types.Clock
}

var (
	errAccountInformationEmpty = errors.New("no account information available")
)

func accountInformationEqual(a, b prosper.AccountInformation) bool {
	if a.AvailableCashBalance != b.AvailableCashBalance {
		return false
	}
	if a.TotalPrincipalReceivedOnActiveNotes != b.TotalPrincipalReceivedOnActiveNotes {
		return false
	}
	if a.OutstandingPrincipalOnActiveNotes != b.OutstandingPrincipalOnActiveNotes {
		return false
	}
	if a.LastWithdrawAmount != b.LastWithdrawAmount {
		return false
	}
	if a.LastDepositAmount != b.LastDepositAmount {
		return false
	}
	if !a.LastDepositDate.Equal(b.LastDepositDate) {
		return false
	}
	if a.PendingInvestmentsPrimaryMarket != b.PendingInvestmentsPrimaryMarket {
		return false
	}
	if a.PendingInvestmentsSecondaryMarket != b.PendingInvestmentsSecondaryMarket {
		return false
	}
	if a.PendingQuickInvestOrders != b.PendingQuickInvestOrders {
		return false
	}
	if a.TotalAmountInvestedOnActiveNotes != b.TotalAmountInvestedOnActiveNotes {
		return false
	}
	if a.TotalAccountValue != b.TotalAccountValue {
		return false
	}
	if a.InflightGross != b.InflightGross {
		return false
	}
	if !a.LastWithdrawDate.Equal(b.LastWithdrawDate) {
		return false
	}
	return true
}

func NewRedisLogger(updates <-chan prosper.AccountInformation) (redisLogger, error) {
	r, err := redis.New()
	if err != nil {
		return redisLogger{}, err
	}
	return redisLogger{
		accountUpdates: updates,
		redis:          r,
		clock:          types.DefaultClock{},
	}, nil
}

func (r redisLogger) Run() {
	last, err := r.getAccountInformation()
	if err != nil && err != errAccountInformationEmpty {
		log.Printf("failed to get account information: %v", err)
	}

	for {
		current, more := <-r.accountUpdates
		if !more {
			return
		}
		if accountInformationEqual(current, last) {
			continue
		}
		log.Printf("new account information: %+v", current)

		record := redis.AccountRecord{
			Value:     current,
			Timestamp: r.clock.Now(),
		}
		if err = r.saveAccountInformation(record); err != nil {
			log.Printf("failed to save account information: %v", err)
		} else {
			last = current
		}
	}
}

func (r redisLogger) getAccountInformation() (prosper.AccountInformation, error) {
	accountSerialized, err := r.redis.LRange(redis.KeyAccountInformation, 0, 0)
	if err != nil {
		return prosper.AccountInformation{}, err
	}
	if len(accountSerialized) < 1 {
		return prosper.AccountInformation{}, errAccountInformationEmpty
	}
	var record redis.AccountRecord
	err = json.Unmarshal([]byte(accountSerialized[0]), &record)
	if err != nil {
		return prosper.AccountInformation{}, err
	}
	return record.Value, nil
}

func (r redisLogger) saveAccountInformation(record redis.AccountRecord) error {
	serialized, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = r.redis.LPush(redis.KeyAccountInformation, string(serialized))
	if err != nil {
		return err
	}
	return nil
}
