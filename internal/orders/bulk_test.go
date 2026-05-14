package orders

import (
	"context"
	"sync/atomic"
	"testing"
)

type fakeStore struct {
	callCount atomic.Int32
	failIDs   map[int32]error
}

func (f *fakeStore) PlaceOrder(_ context.Context, _ int, req PlaceOrderRequest) (Order, bool, error) {
	f.callCount.Add(1)
	for _, item := range req.Items {
		if err, shouldFail := f.failIDs[item.ProductID]; shouldFail {
			return Order{}, false, err
		}
	}
	return Order{ID: 1}, true, nil
}

func TestProcessBulkOrder_AllSucceed(t *testing.T) {
	fake := &fakeStore{}
	req := BulkOrderRequest{
		Items: []BulkOrderItem{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
			{ProductID: 3, Quantity: 5},
		},
	}

	resp, err := processBulkWithPlacer(context.Background(), fake, 42, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Succeeded != 3 {
		t.Errorf("expected 3 succeeded, got %d", resp.Succeeded)
	}
	if resp.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", resp.Failed)
	}
	if fake.callCount.Load() != 3 {
		t.Errorf("expected PlaceOrder to be called 3 times, got %d", fake.callCount.Load())
	}
}

func TestProcessBulkOrder_PartialFailure(t *testing.T) {
	fake := &fakeStore{
		failIDs: map[int32]error{
			2: ErrInsufficientStock,
		},
	}
	req := BulkOrderRequest{
		Items: []BulkOrderItem{
			{ProductID: 1, Quantity: 1},
			{ProductID: 2, Quantity: 1},
			{ProductID: 3, Quantity: 1},
		},
	}

	resp, err := processBulkWithPlacer(context.Background(), fake, 42, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Succeeded != 2 {
		t.Errorf("expected 2 succeeded, got %d", resp.Succeeded)
	}
	if resp.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", resp.Failed)
	}

	// Verificamos que el resultado fallido tiene la razón correcta
	var failedResult *BulkOrderResult
	for i, r := range resp.Results {
		if r.Status == "failed" {
			failedResult = &resp.Results[i]
		}
	}
	if failedResult == nil {
		t.Fatal("expected a failed result but found none")
	}
	if failedResult.ProductID != 2 {
		t.Errorf("expected failed productID 2, got %d", failedResult.ProductID)
	}
}

func TestProcessBulkOrder_AllFail(t *testing.T) {
	fake := &fakeStore{
		failIDs: map[int32]error{
			1: ErrProductNotFound,
			2: ErrInsufficientStock,
		},
	}
	req := BulkOrderRequest{
		Items: []BulkOrderItem{
			{ProductID: 1, Quantity: 1},
			{ProductID: 2, Quantity: 1},
		},
	}

	resp, err := processBulkWithPlacer(context.Background(), fake, 42, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Succeeded != 0 {
		t.Errorf("expected 0 succeeded, got %d", resp.Succeeded)
	}
	if resp.Failed != 2 {
		t.Errorf("expected 2 failed, got %d", resp.Failed)
	}
}

func TestProcessBulkOrder_Concurrency(t *testing.T) {
	fake := &fakeStore{}

	items := make([]BulkOrderItem, 50)
	for i := range items {
		items[i] = BulkOrderItem{ProductID: int32(i + 1), Quantity: 1}
	}

	req := BulkOrderRequest{Items: items}
	resp, err := processBulkWithPlacer(context.Background(), fake, 1, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Succeeded != 50 {
		t.Errorf("expected 50 succeeded, got %d", resp.Succeeded)
	}
	if fake.callCount.Load() != 50 {
		t.Errorf("expected 50 calls, got %d", fake.callCount.Load())
	}
}

func TestProcessBulkOrder_EmptyItems(t *testing.T) {
	fake := &fakeStore{}
	req := BulkOrderRequest{Items: []BulkOrderItem{}}

	resp, err := processBulkWithPlacer(context.Background(), fake, 1, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Succeeded != 0 || resp.Failed != 0 {
		t.Errorf("expected empty response, got %+v", resp)
	}
	if fake.callCount.Load() != 0 {
		t.Error("expected no calls to PlaceOrder for empty request")
	}
}
