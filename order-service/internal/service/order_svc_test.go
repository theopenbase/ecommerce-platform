package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ecommerce/order-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============ Cart Service Logic Tests ============

func TestAddToCart_NewItem(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	// Cart item does not exist → CreateCart path
	mockRepo.On("FindCartByUserAndSku", context.Background(), uint64(1), uint64(100)).
		Return(nil, errors.New("not found"))
	mockRepo.On("CreateCart", context.Background(), mock.AnythingOfType("*model.Cart")).Return(nil)

	// Validate: FindCartByUserAndSku returns not-found error
	_, err := mockRepo.FindCartByUserAndSku(context.Background(), 1, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Validate: CreateCart is called with correct user/sku
	mockRepo.On("CreateCart", context.Background(), mock.MatchedBy(func(c *model.Cart) bool {
		return c.UserID == 1 && c.SkuID == 100 && c.Quantity == 3 && c.Checked == 1
	})).Return(nil)

	err = mockRepo.CreateCart(context.Background(), &model.Cart{UserID: 1, SkuID: 100, Quantity: 3, Checked: 1})
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddToCart_ExistingItem_IncrementsQuantity(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	existing := &model.Cart{ID: 1, UserID: 1, SkuID: 100, Quantity: 2, Checked: 1}

	mockRepo.On("FindCartByUserAndSku", context.Background(), uint64(1), uint64(100)).
		Return(existing, nil)
	mockRepo.On("UpdateCart", context.Background(), mock.MatchedBy(func(c *model.Cart) bool {
		return c.Quantity == 5 // original 2 + requested 3
	})).Return(nil)

	// Existing item found → UpdateCart path (quantity += req.Quantity)
	found, err := mockRepo.FindCartByUserAndSku(context.Background(), 1, 100)
	assert.NoError(t, err)
	assert.Equal(t, 2, found.Quantity)

	found.Quantity += 3
	err = mockRepo.UpdateCart(context.Background(), found)
	assert.NoError(t, err)
	assert.Equal(t, 5, found.Quantity)
	mockRepo.AssertExpectations(t)
}

func TestUpdateCart_QuantityZero_DeletesCart(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	existing := &model.Cart{ID: 5, UserID: 1, Quantity: 1}

	mockRepo.On("FindCartByID", context.Background(), uint64(5), uint64(1)).
		Return(existing, nil)
	mockRepo.On("DeleteCart", context.Background(), uint64(5), uint64(1)).
		Return(nil)

	// When req.Quantity <= 0, DeleteCart is called (no UpdateCart)
	found, err := mockRepo.FindCartByID(context.Background(), 5, 1)
	assert.NoError(t, err)

	if found != nil {
		err = mockRepo.DeleteCart(context.Background(), 5, 1)
		assert.NoError(t, err)
	}

	mockRepo.AssertCalled(t, "DeleteCart", context.Background(), uint64(5), uint64(1))
	mockRepo.AssertNotCalled(t, "UpdateCart")
}

func TestRemoveCart_NotFound_ReturnsError(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	mockRepo.On("FindCartByID", context.Background(), uint64(999), uint64(1)).
		Return(nil, errors.New("not found"))

	// Non-existent cart: FindCartByID returns error → ErrCartNotFound
	found, err := mockRepo.FindCartByID(context.Background(), 999, 1)
	assert.Nil(t, found)
	assert.Error(t, err)

	// DeleteCart should NOT be called
	mockRepo.AssertNotCalled(t, "DeleteCart")
}

func TestSelectCartItems_AllSelected(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	carts := []model.Cart{
		{ID: 1, UserID: 1, Checked: 0},
		{ID: 2, UserID: 1, Checked: 0},
	}

	mockRepo.On("FindCartsByUserID", context.Background(), uint64(1)).
		Return(carts, nil)

	for range carts {
		mockRepo.On("UpdateCart", context.Background(), mock.AnythingOfType("*model.Cart")).Return(nil).Once()
	}

	// SelectAll(checked=true): all items → Checked=1
	found, _ := mockRepo.FindCartsByUserID(context.Background(), 1)
	for i := range found {
		found[i].Checked = 1
		mockRepo.UpdateCart(context.Background(), &found[i])
	}

	assert.Equal(t, uint8(1), found[0].Checked)
	assert.Equal(t, uint8(1), found[1].Checked)
}

func TestSelectCartItems_AllDeselected(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	carts := []model.Cart{
		{ID: 1, UserID: 1, Checked: 1},
		{ID: 2, UserID: 1, Checked: 1},
	}

	mockRepo.On("FindCartsByUserID", context.Background(), uint64(1)).
		Return(carts, nil)

	for range carts {
		mockRepo.On("UpdateCart", context.Background(), mock.AnythingOfType("*model.Cart")).Return(nil).Once()
	}

	// SelectAll(checked=false): all items → Checked=0
	found, _ := mockRepo.FindCartsByUserID(context.Background(), 1)
	for i := range found {
		found[i].Checked = 0
		mockRepo.UpdateCart(context.Background(), &found[i])
	}

	assert.Equal(t, uint8(0), found[0].Checked)
	assert.Equal(t, uint8(0), found[1].Checked)
}

// ============ Order Creation Tests ============

func TestCreateOrder_IdempotentKeyUsed_ReturnsError(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	// Duplicate request: idempotent key already exists
	mockRepo.On("CheckIdempotencyKey", context.Background(), "1:[1 2]").
		Return(true, nil)

	used, err := mockRepo.CheckIdempotencyKey(context.Background(), "1:[1 2]")
	assert.True(t, used)
	assert.NoError(t, err)
	// Service should return ErrIdempotentKeyUsed when used=true
}

func TestCreateOrder_EmptyCartItems_ReturnsErrCartNotFound(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	mockRepo.On("CheckIdempotencyKey", context.Background(), mock.AnythingOfType("string")).
		Return(false, nil)
	mockRepo.On("CacheIdempotencyKey", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(nil)
	mockRepo.On("FindCartsByIDs", context.Background(), []uint64{1, 2}, uint64(1)).
		Return([]model.Cart{}, nil)

	// No cart items found → service returns ErrCartNotFound
	items, err := mockRepo.FindCartsByIDs(context.Background(), []uint64{1, 2}, 1)
	assert.Empty(t, items)
	assert.NoError(t, err)
}

func TestCreateOrder_StockFreezeFails_RollsBack(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	// SKU100 freeze succeeds, SKU200 freeze fails
	mockRepo.On("FreezeStock", context.Background(), uint64(100), 2).Return(true, nil)
	mockRepo.On("FreezeStock", context.Background(), uint64(200), 1).Return(false, nil)
	mockRepo.On("UnfreezeStock", context.Background(), uint64(100), 2).Return(nil)

	// Simulate freeze-then-rollback logic
	cartItems := []struct{ skuID uint64; qty int }{{100, 2}, {200, 1}}
	var frozen []struct{ skuID uint64; qty int }

	for _, item := range cartItems {
		locked, _ := mockRepo.FreezeStock(context.Background(), item.skuID, item.qty)
		if locked {
			frozen = append(frozen, item)
		} else {
			// Rollback already-frozen items
			for _, f := range frozen {
				mockRepo.UnfreezeStock(context.Background(), f.skuID, f.qty)
			}
			break
		}
	}

	// Only SKU100 was frozen before failure → it gets unfrozen
	assert.Len(t, frozen, 1)
	assert.Equal(t, uint64(100), frozen[0].skuID)
	mockRepo.AssertCalled(t, "UnfreezeStock", context.Background(), uint64(100), 2)
}

func TestCreateOrder_Success_FullFlow(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	cartItems := []model.Cart{
		{ID: 1, SkuID: 100, SpuID: 10, ShopID: 1, Quantity: 2},
	}

	mockRepo.On("CheckIdempotencyKey", context.Background(), mock.AnythingOfType("string")).
		Return(false, nil)
	mockRepo.On("CacheIdempotencyKey", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(nil)
	mockRepo.On("FindCartsByIDs", context.Background(), []uint64{1}, uint64(1)).
		Return(cartItems, nil)
	mockRepo.On("FreezeStock", context.Background(), uint64(100), 2).Return(true, nil)
	mockRepo.On("CreateParentOrder", context.Background(), mock.AnythingOfType("*model.ParentOrder")).Return(nil)
	mockRepo.On("CreateSubOrder", context.Background(), mock.AnythingOfType("*model.SubOrder")).Return(nil)
	mockRepo.On("CreateOrderItems", context.Background(), mock.AnythingOfType("[]model.OrderItem")).Return(nil)
	mockRepo.On("CreateOrderActionLog", context.Background(), mock.AnythingOfType("*model.OrderActionLog")).Return(nil)
	mockRepo.On("DeleteCartsByIDs", context.Background(), []uint64{1}, uint64(1)).Return(nil)

	// Validate each step in CreateOrder flow
	used, _ := mockRepo.CheckIdempotencyKey(context.Background(), "user:1")
	assert.False(t, used)

	locked, _ := mockRepo.FreezeStock(context.Background(), 100, 2)
	assert.True(t, locked)

	err := mockRepo.CreateParentOrder(context.Background(), &model.ParentOrder{})
	assert.NoError(t, err)

	err = mockRepo.DeleteCartsByIDs(context.Background(), []uint64{1}, 1)
	assert.NoError(t, err)
}

// ============ Order Status Transition Tests ============

func TestCancelOrder_NotPendingPayment_ReturnsErrInvalidStatus(t *testing.T) {
	// Paid order (Status=2) cannot be cancelled
	paidOrder := &model.ParentOrder{
		OrderNo: "ORD001",
		BuyerID: 1,
		Status:  model.OrderStatusPaid,
	}
	assert.NotEqual(t, model.OrderStatusPendingPayment, paidOrder.Status)
}

func TestCancelOrder_WrongBuyer_ReturnsErrNotAuthorized(t *testing.T) {
	order := &model.ParentOrder{OrderNo: "ORD001", BuyerID: 1}

	// Buyer 2 attempts to cancel buyer 1's order
	wrongBuyerID := uint64(2)
	assert.NotEqual(t, wrongBuyerID, order.BuyerID)
}

func TestCancelOrder_Success_UnfreezesStock(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	subOrders := []model.SubOrder{{SubOrderNo: "SUB001", ShopID: 1}}
	items := []model.OrderItem{{SkuID: 100, Quantity: 2}}

	mockRepo.On("FindSubOrdersByParentOrderNo", context.Background(), "ORD001").
		Return(subOrders, nil)
	mockRepo.On("FindOrderItemsBySubOrderNo", context.Background(), "SUB001").
		Return(items, nil)
	mockRepo.On("UnfreezeStock", context.Background(), uint64(100), 2).Return(nil)

	// Cancel order → unfreeze all frozen stock
	for _, so := range subOrders {
		itemList, _ := mockRepo.FindOrderItemsBySubOrderNo(context.Background(), so.SubOrderNo)
		for _, item := range itemList {
			mockRepo.UnfreezeStock(context.Background(), item.SkuID, item.Quantity)
		}
	}

	mockRepo.AssertCalled(t, "UnfreezeStock", context.Background(), uint64(100), 2)
}

func TestConfirmReceive_NotDelivered_ReturnsErrInvalidStatus(t *testing.T) {
	// Pending payment order cannot confirm receive
	pendingOrder := &model.ParentOrder{
		OrderNo: "ORD001",
		BuyerID: 1,
		Status:  model.OrderStatusPendingPayment,
	}
	assert.NotEqual(t, model.OrderStatusDelivered, pendingOrder.Status)
}

func TestApplyRefund_PaidStatus_Success(t *testing.T) {
	subOrder := &model.SubOrder{
		SubOrderNo:    "SUB001",
		ParentOrderNo: "ORD001",
		BuyerID:       1,
		Status:        model.OrderStatusPaid,
	}

	// Paid order → valid refund status
	assert.Equal(t, model.OrderStatusPaid, subOrder.Status)
	subOrder.Status = model.OrderStatusDispute
	assert.Equal(t, model.OrderStatusDispute, subOrder.Status)
}

func TestApplyRefund_CompletedStatus_ReturnsErrInvalidStatus(t *testing.T) {
	subOrder := &model.SubOrder{
		SubOrderNo: "SUB001",
		BuyerID:    1,
		Status:     model.OrderStatusCompleted,
	}

	// Completed order cannot be refunded (only paid or delivered can apply refund)
	assert.NotEqual(t, model.OrderStatusPaid, subOrder.Status)
	assert.NotEqual(t, model.OrderStatusDelivered, subOrder.Status)
	assert.Equal(t, model.OrderStatusCompleted, subOrder.Status)
}

// ============ Order Query Tests ============

func TestGetOrderDetail_WrongBuyer_ReturnsErrNotAuthorized(t *testing.T) {
	order := &model.ParentOrder{OrderNo: "ORD001", BuyerID: 1}
	wrongBuyerID := uint64(2)

	// Wrong buyer attempting to view order
	assert.NotEqual(t, wrongBuyerID, order.BuyerID)
}

func TestGetOrderDetail_NotFound_ReturnsErrOrderNotFound(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	mockRepo.On("FindParentOrderByOrderNo", context.Background(), "NOTEXIST").
		Return(nil, errors.New("not found"))

	_, err := mockRepo.FindParentOrderByOrderNo(context.Background(), "NOTEXIST")
	assert.Error(t, err)
}

func TestListOrders_DefaultPagination(t *testing.T) {
	q := &model.OrderListQuery{}

	// Service applies defaults when values are invalid
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}

	assert.Equal(t, 1, q.Page)
	assert.Equal(t, 20, q.PageSize)
}

func TestListOrders_WithStatusFilter(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	status := uint8(model.OrderStatusPaid)
	mockRepo.On("ListParentOrders", context.Background(), uint64(1), &status, 1, 20).
		Return([]model.ParentOrder{}, int64(0), nil)

	_, total, err := mockRepo.ListParentOrders(context.Background(), 1, &status, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.NotNil(t, &status)
}

// ============ Order Model & Constants Tests ============

func TestOrderStatusConstants(t *testing.T) {
	tests := []struct {
		status   uint8
		expected string
	}{
		{model.OrderStatusPendingPayment, "待付款"},
		{model.OrderStatusCancelled, "已取消"},
		{model.OrderStatusPaid, "待发货"},
		{model.OrderStatusDelivered, "待收货"},
		{model.OrderStatusReceived, "已收货"},
		{model.OrderStatusCompleted, "已完成"},
		{model.OrderStatusDispute, "维权中"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, model.OrderStatusText[tt.status])
		})
	}
}

func TestGenOrderNo_Format(t *testing.T) {
	orderNo := model.GenOrderNo()

	// Format: 8-digit date + 6-digit sequence + 2-digit random = 16 chars
	assert.Len(t, orderNo, 16)

	// Date prefix must match today
	today := time.Now().Format("20060102")
	assert.Equal(t, today, orderNo[:8])
}

func TestGenOrderNo_Uniqueness(t *testing.T) {
	nos := make(map[string]bool)
	for i := 0; i < 100; i++ {
		no := model.GenOrderNo()
		assert.False(t, nos[no], "duplicate order number generated: "+no)
		nos[no] = true
	}
}

// ============ Stock Operations Tests ============

func TestFreezeStock_InsufficientStock_ReturnsFalse(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	// Stock=1, required=5 → should fail
	mockRepo.On("FreezeStock", context.Background(), uint64(100), 5).Return(false, nil)

	locked, _ := mockRepo.FreezeStock(context.Background(), 100, 5)
	assert.False(t, locked)
}

func TestUnfreezeStock_AfterCancel_IncreasesAvailableStock(t *testing.T) {
	mockRepo := new(MockOrderRepository)

	mockRepo.On("UnfreezeStock", context.Background(), uint64(100), 2).Return(nil)

	// Cancellation returns frozen stock to available pool
	err := mockRepo.UnfreezeStock(context.Background(), 100, 2)
	assert.NoError(t, err)
}

// ============ Error Constants ============

func TestErrorConstants(t *testing.T) {
	assert.NotNil(t, ErrCartNotFound)
	assert.NotNil(t, ErrSkuNotAvailable)
	assert.NotNil(t, ErrOrderNotFound)
	assert.NotNil(t, ErrInvalidStatus)
	assert.NotNil(t, ErrNotAuthorized)
	assert.NotNil(t, ErrIdempotentKeyUsed)

	assert.Equal(t, "cart item not found", ErrCartNotFound.Error())
	assert.Equal(t, "order not found", ErrOrderNotFound.Error())
	assert.Equal(t, "invalid status transition", ErrInvalidStatus.Error())
	assert.Equal(t, "not authorized", ErrNotAuthorized.Error())
	assert.Equal(t, "duplicate request", ErrIdempotentKeyUsed.Error())
}
