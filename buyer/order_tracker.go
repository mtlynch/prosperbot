package buyer

import (
	"log"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type orderStatusQueryWorker struct {
	querier      prosper.OrderStatusQuerier
	orderUpdates chan<- prosper.OrderResponse
}

func (qw orderStatusQueryWorker) QueryUntilComplete(orderID prosper.OrderID) {
	retries := 3
	for {
		if retries == 0 {
			return
		}
		response, err := qw.querier.OrderStatus(orderID)
		if err != nil {
			log.Printf("Failed to query orderStatus for %v, err: %v", orderID, err)
			retries -= 1
			continue
		}
		go func() { qw.orderUpdates <- response }()

		if response.OrderStatus == prosper.OrderCompleted || response.BidStatus[0].Result != prosper.NoBidResult {
			log.Printf("order %v is complete: %v", orderID, response)
			return
		}
	}
}

type orderTracker struct {
	querier      prosper.OrderStatusQuerier
	orders       <-chan prosper.OrderID
	orderUpdates chan<- prosper.OrderResponse
}

func (ot orderTracker) Run() {
	for {
		orderID := <-ot.orders
		log.Printf("new order: %v", orderID)
		worker := orderStatusQueryWorker{ot.querier, ot.orderUpdates}
		go worker.QueryUntilComplete(orderID)
	}
}
