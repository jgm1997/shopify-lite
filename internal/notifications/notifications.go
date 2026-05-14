package notifications

import (
	"context"
	"log"
	"sync"
	"time"
)

type Notifier struct {
	wg sync.WaitGroup
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

type BulkOrderSummary struct {
	CustomerID int
	Succeeded  int
	Failed     int
}

func (n *Notifier) SendBulkOrderSummary(summary BulkOrderSummary) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		select {
		case <-time.After(2 * time.Second):
			log.Printf("[notify] customer %d - bulk order: %d succeeded %d failed", summary.CustomerID, summary.Succeeded, summary.Failed)
		case <-ctx.Done():
			log.Printf("[notify] timeout sending notification to customer %d", summary.CustomerID)
		}
	}()
}

func (n *Notifier) Shutdown() {
	n.wg.Wait()
}
