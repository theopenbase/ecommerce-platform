package service

import (
	"context"
	"log"
	"time"

	"github.com/ecommerce/order-service/internal/model"
	"github.com/ecommerce/order-service/internal/repository"
)

// StockCompensationTask 超卖防护定时任务
// 定期扫描超时未支付的冻结记录，归还库存，防止库存泄露
type StockCompensationTask struct {
	repo      *repository.OrderRepository
	timeout   time.Duration
	interval  time.Duration
	stopCh    chan struct{}
}

func NewStockCompensationTask(repo *repository.OrderRepository, timeout, interval time.Duration) *StockCompensationTask {
	return &StockCompensationTask{
		repo:     repo,
		timeout:  timeout,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动补偿任务（阻塞）
func (t *StockCompensationTask) Start(ctx context.Context) {
	log.Printf("[StockCompensation] started: timeout=%v, interval=%v", t.timeout, t.interval)
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[StockCompensation] stopped: context cancelled")
			return
		case <-t.stopCh:
			log.Printf("[StockCompensation] stopped: stop signal")
			return
		case <-ticker.C:
			t.compensate(ctx)
		}
	}
}

// Stop 停止补偿任务
func (t *StockCompensationTask) Stop() {
	close(t.stopCh)
}

// compensate 执行一次补偿扫描
func (t *StockCompensationTask) compensate(ctx context.Context) {
	records, err := t.repo.FindExpiredFrozenStocks(ctx, t.timeout)
	if err != nil {
		log.Printf("[StockCompensation] scan error: %v", err)
		return
	}

	if len(records) == 0 {
		return
	}

	log.Printf("[StockCompensation] found %d expired frozen stock records", len(records))

	for _, record := range records {
		// 1. 解冻 Redis 库存
		if err := t.repo.UnfreezeStock(ctx, record.SkuID, record.Quantity); err != nil {
			log.Printf("[StockCompensation] unfreeze stock error: sku=%d, order=%s, err=%v",
				record.SkuID, record.OrderNo, err)
			continue
		}

		// 2. 更新冻结记录状态为已回滚
		if err := t.repo.UpdateFrozenStockState(ctx, record.ID, repository.FrozenStateRolled); err != nil {
			log.Printf("[StockCompensation] update state error: id=%d, err=%v", record.ID, err)
		}

		log.Printf("[StockCompensation] rolled back: order=%s, sku=%d, qty=%d",
			record.OrderNo, record.SkuID, record.Quantity)
	}
}

// RunOnce 用于手动触发一次补偿（测试用）
func (t *StockCompensationTask) RunOnce(ctx context.Context) ([]model.FrozenStock, error) {
	records, err := t.repo.FindExpiredFrozenStocks(ctx, t.timeout)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		t.repo.UnfreezeStock(ctx, record.SkuID, record.Quantity)
		t.repo.UpdateFrozenStockState(ctx, record.ID, repository.FrozenStateRolled)
	}
	return records, nil
}
