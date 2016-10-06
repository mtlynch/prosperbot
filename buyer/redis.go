package buyer

import (
	"encoding/json"
	"log"

	"github.com/mtlynch/gofn-prosper/prosper"

	"github.com/mtlynch/prosperbot/clock"
	"github.com/mtlynch/prosperbot/redis"
)

type orderStatusLogger struct {
	redis        redis.RedisSetter
	orderUpdates <-chan prosper.OrderResponse
	done         chan<- bool
	clock        clock.Clock
}

func NewOrderStatusLogger(orderUpdates <-chan prosper.OrderResponse) (orderStatusLogger, error) {
	r, err := redis.New()
	if err != nil {
		return orderStatusLogger{}, err
	}
	done := make(chan bool)
	return orderStatusLogger{
		redis:        r,
		orderUpdates: orderUpdates,
		done:         done,
		clock:        clock.DefaultClock{},
	}, nil
}

func (r orderStatusLogger) Run() {
	for {
		order, more := <-r.orderUpdates
		if !more {
			r.done <- true
			return
		}
		log.Printf("new order update: %+v", order)

		record := redis.OrderRecord{
			Order:     order,
			Timestamp: r.clock.Now(),
		}
		if err := r.saveOrderStatus(record); err != nil {
			log.Printf("failed to save order status: %v", err)
		}
	}
}

func (r orderStatusLogger) saveOrderStatus(record redis.OrderRecord) error {
	serialized, err := json.Marshal(record)
	if err != nil {
		return err
	}
	key := redis.KeyPrefixOrders + string(record.Order.OrderID)
	_, err = r.redis.Set(key, string(serialized))
	if err != nil {
		return err
	}
	return nil
}
