package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ecommerce/order-service/internal/model"
	"github.com/ecommerce/order-service/internal/pkg/goodsclient"
	"github.com/ecommerce/order-service/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrCartNotFound      = errors.New("cart item not found")
	ErrSkuNotAvailable   = errors.New("sku not available or insufficient stock")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidStatus     = errors.New("invalid status transition")
	ErrNotAuthorized     = errors.New("not authorized")
	ErrIdempotentKeyUsed = errors.New("duplicate request")
)

type OrderService struct {
	repo      *repository.OrderRepository
	goodsCli  *goodsclient.GoodsClient
}

func NewOrderService(repo *repository.OrderRepository, goodsCli *goodsclient.GoodsClient) *OrderService {
	return &OrderService{repo: repo, goodsCli: goodsCli}
}

// ============ 购物车 ============

// AddToCart 添加到购物车
func (s *OrderService) AddToCart(ctx context.Context, userID uint64, req *model.CartAddRequest) (*model.Cart, error) {
	// 通过 goods-service RPC 获取 SKU 快照，填充商品信息
	snapshot, err := s.goodsCli.GetSkuSnapshot(ctx, req.SkuID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku info: %w", err)
	}

	existing, err := s.repo.FindCartByUserAndSku(ctx, userID, req.SkuID)
	if err == nil && existing != nil {
		existing.Quantity += req.Quantity
		// 快照以首次添加时为准，不随价格变动更新
		if err := s.repo.UpdateCart(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	cart := &model.Cart{
		UserID:    userID,
		SkuID:     req.SkuID,
		SpuID:     snapshot.SpuID,
		ShopID:    snapshot.ShopID,
		Quantity:  req.Quantity,
		Checked:   1,
		SkuCode:   snapshot.SkuCode,
		Title:     snapshot.Title,
		SkuAttrs:   snapshot.SkuAttrs,
		PriceTag:  snapshot.PriceTag,
		PriceSell: snapshot.PriceSell,
	}
	if err := s.repo.CreateCart(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// UpdateCart 更新购物车数量
func (s *OrderService) UpdateCart(ctx context.Context, userID, cartID uint64, req *model.CartUpdateRequest) (*model.Cart, error) {
	cart, err := s.repo.FindCartByID(ctx, cartID, userID)
	if err != nil {
		return nil, ErrCartNotFound
	}
	if req.Quantity <= 0 {
		if err := s.repo.DeleteCart(ctx, cartID, userID); err != nil {
			return nil, err
		}
		return nil, nil
	}
	cart.Quantity = req.Quantity
	if err := s.repo.UpdateCart(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// RemoveCart 删除购物车项
func (s *OrderService) RemoveCart(ctx context.Context, userID, cartID uint64) error {
	_, err := s.repo.FindCartByID(ctx, cartID, userID)
	if err != nil {
		return ErrCartNotFound
	}
	return s.repo.DeleteCart(ctx, cartID, userID)
}

// GetCart 获取用户购物车
func (s *OrderService) GetCart(ctx context.Context, userID uint64) (*model.CartResponse, error) {
	carts, err := s.repo.FindCartsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var items []model.CartItemResponse
	for _, cart := range carts {
		items = append(items, model.CartItemResponse{
			CartID:   cart.ID,
			SkuID:    cart.SkuID,
			SpuID:    cart.SpuID,
			ShopID:   cart.ShopID,
			Quantity: cart.Quantity,
			Checked:  cart.Checked == 1,
		})
	}
	return &model.CartResponse{Items: items}, nil
}

// SelectCartItems 全选/取消全选
func (s *OrderService) SelectCartItems(ctx context.Context, userID uint64, checked bool) error {
	carts, err := s.repo.FindCartsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	checkVal := uint8(0)
	if checked {
		checkVal = 1
	}
	for _, cart := range carts {
		cart.Checked = checkVal
		s.repo.UpdateCart(ctx, &cart)
	}
	return nil
}

// ============ 订单创建 ============

// CreateOrder 创建订单（带库存冻结和幂等保证）
func (s *OrderService) CreateOrder(ctx context.Context, userID uint64, req *model.ConfirmOrderRequest) (*model.OrderResponse, error) {
	idempotentKey := fmt.Sprintf("%d:%v", userID, req.Items)
	used, _ := s.repo.CheckIdempotencyKey(ctx, idempotentKey)
	if used {
		return nil, ErrIdempotentKeyUsed
	}
	s.repo.CacheIdempotencyKey(ctx, idempotentKey, 10*time.Minute)

	cartItems, err := s.repo.FindCartsByIDs(ctx, req.Items, userID)
	if err != nil || len(cartItems) == 0 {
		return nil, ErrCartNotFound
	}

	// 按店铺分组
	shopGroups := make(map[uint64][]model.Cart)
	for _, item := range cartItems {
		shopGroups[item.ShopID] = append(shopGroups[item.ShopID], item)
	}

	// 冻结每个 SKU 的库存（Redis 原子操作）
	for _, item := range cartItems {
		locked, err := s.repo.FreezeStock(ctx, item.SkuID, item.Quantity)
		if err != nil || !locked {
			// 回滚已冻结的库存
			for _, prev := range cartItems {
				if prev.SkuID == item.SkuID {
					break
				}
				s.repo.UnfreezeStock(ctx, prev.SkuID, prev.Quantity)
			}
			return nil, ErrSkuNotAvailable
		}
	}

	parentOrderNo := model.GenOrderNo()

	// 写入冻结记录（用于超时补偿）
	for _, item := range cartItems {
		record := &model.FrozenStock{
			SkuID:    item.SkuID,
			OrderNo:  parentOrderNo,
			Quantity: item.Quantity,
			State:    repository.FrozenStateActive,
		}
		if err := s.repo.CreateFrozenStock(ctx, record); err != nil {
			// 回滚已冻结的库存
			for _, item := range cartItems {
				s.repo.UnfreezeStock(ctx, item.SkuID, item.Quantity)
			}
			return nil, err
		}
	}

	// 组装店铺分组数据
	type shopGroup struct {
		shopID     uint64
		items      []model.Cart
		subOrderNo string
		freight    float64
		discount   float64
		payAmount  float64
	}
	var groups []shopGroup
	var totalAmount, totalFreight, totalDiscount float64

	for shopID, items := range shopGroups {
		var shopTotal float64
		for _, c := range items {
			shopTotal += float64(c.Quantity) * c.PriceSell
		}
		grp := shopGroup{
			shopID:     shopID,
			items:      items,
			subOrderNo: model.GenOrderNo(),
			freight:    10.0,
			discount:   0,
			payAmount:  shopTotal + 10.0,
		}
		totalAmount += shopTotal
		totalFreight += grp.freight
		totalDiscount += grp.discount
		groups = append(groups, grp)
	}
	payAmount := totalAmount + totalFreight - totalDiscount

	parentOrder := &model.ParentOrder{
		OrderNo:       parentOrderNo,
		BuyerID:       userID,
		TotalAmount:   totalAmount,
		FreightAmount: totalFreight,
		DiscountAmt:   totalDiscount,
		PayAmount:     payAmount,
		Status:        model.OrderStatusPendingPayment,
	}

	// DB 事务：原子创建父订单+子订单+订单项+操作日志+清空购物车
	err = s.repo.WithTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(parentOrder).Error; err != nil {
			return err
		}

		var subOrders []model.SubOrderInfo
		for _, grp := range groups {
			subOrder := &model.SubOrder{
				SubOrderNo:    grp.subOrderNo,
				ParentOrderNo: parentOrderNo,
				BuyerID:       userID,
				ShopID:        grp.shopID,
				Status:        model.OrderStatusPendingPayment,
				FreightAmount: grp.freight,
				DiscountAmt:   grp.discount,
				PayAmount:     grp.payAmount,
			}
			if err := tx.Create(subOrder).Error; err != nil {
				return err
			}

			var orderItems []model.OrderItem
			for _, cartItem := range grp.items {
				item := model.OrderItem{
					SubOrderNo: grp.subOrderNo,
					SkuID:      cartItem.SkuID,
					SpuID:      cartItem.SpuID,
					SkuCode:    cartItem.SkuCode,
					Title:      cartItem.Title,
					SkuAttrs:   cartItem.SkuAttrs,
					PriceTag:   cartItem.PriceTag,
					PriceSell:  cartItem.PriceSell,
					Quantity:   cartItem.Quantity,
					ItemTotal:  float64(cartItem.Quantity) * cartItem.PriceSell,
				}
				orderItems = append(orderItems, item)
			}
			if err := tx.Create(&orderItems).Error; err != nil {
				return err
			}
			subOrders = append(subOrders, model.SubOrderInfo{
				SubOrder: subOrder,
				Items:    orderItems,
				ShopName: "",
			})
		}

		if err := tx.Create(&model.OrderActionLog{
			OrderNo:  parentOrderNo,
			Action:   "create",
			Operator: fmt.Sprintf("%d", userID),
			Note:     "订单创建",
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("id IN ? AND user_id = ?", req.Items, userID).Delete(&model.Cart{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// 事务失败：回滚 Redis 冻结库存
		for _, item := range cartItems {
			s.repo.UnfreezeStock(ctx, item.SkuID, item.Quantity)
		}
		return nil, err
	}

	return &model.OrderResponse{
		ParentOrder: parentOrder,
		SubOrders:   nil, // 已在上面填充
		Address:     nil,
	}, nil
}

// ============ 订单状态变更 ============

// Pay 支付成功回调
func (s *OrderService) Pay(ctx context.Context, orderNo string) error {
	order, err := s.repo.FindParentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status != model.OrderStatusPendingPayment {
		return ErrInvalidStatus
	}
	now := time.Now()
	order.Status = model.OrderStatusPaid
	order.PayTime = &now
	if err := s.repo.UpdateParentOrder(ctx, order); err != nil {
		return err
	}
	subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, orderNo)
	for _, so := range subOrders {
		so.Status = model.OrderStatusPaid
		s.repo.UpdateSubOrder(ctx, &so)
	}
	s.repo.CreateOrderActionLog(ctx, &model.OrderActionLog{
		OrderNo: orderNo, Action: "pay", Operator: "system", Note: "支付成功",
	})
	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, buyerID uint64, orderNo, reason string) error {
	order, err := s.repo.FindParentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.BuyerID != buyerID {
		return ErrNotAuthorized
	}
	if order.Status != model.OrderStatusPendingPayment {
		return ErrInvalidStatus
	}
	now := time.Now()
	order.Status = model.OrderStatusCancelled
	order.CancelTime = &now
	order.CancelReason = reason
	if err := s.repo.UpdateParentOrder(ctx, order); err != nil {
		return err
	}
	subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, orderNo)
	for _, so := range subOrders {
		items, _ := s.repo.FindOrderItemsBySubOrderNo(ctx, so.SubOrderNo)
		for _, item := range items {
			s.repo.UnfreezeStock(ctx, item.SkuID, item.Quantity)
		}
	}
	s.repo.CreateOrderActionLog(ctx, &model.OrderActionLog{
		OrderNo: orderNo, Action: "cancel", Operator: fmt.Sprintf("%d", buyerID), Note: "用户取消：" + reason,
	})
	return nil
}

// ConfirmReceive 确认收货
func (s *OrderService) ConfirmReceive(ctx context.Context, buyerID uint64, orderNo string) error {
	order, err := s.repo.FindParentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.BuyerID != buyerID {
		return ErrNotAuthorized
	}
	if order.Status != model.OrderStatusDelivered {
		return ErrInvalidStatus
	}
	now := time.Now()
	order.Status = model.OrderStatusReceived
	order.ReceiveTime = &now
	if err := s.repo.UpdateParentOrder(ctx, order); err != nil {
		return err
	}
	subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, orderNo)
	for _, so := range subOrders {
		so.Status = model.OrderStatusReceived
		s.repo.UpdateSubOrder(ctx, &so)
	}
	s.repo.CreateOrderActionLog(ctx, &model.OrderActionLog{
		OrderNo: orderNo, Action: "receive", Operator: fmt.Sprintf("%d", buyerID), Note: "确认收货",
	})
	return nil
}

// AutoComplete 超时自动完成（定时任务）
func (s *OrderService) AutoComplete(ctx context.Context, orderNo string) error {
	order, err := s.repo.FindParentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status != model.OrderStatusReceived {
		return nil
	}
	now := time.Now()
	order.Status = model.OrderStatusCompleted
	order.FinishTime = &now
	if err := s.repo.UpdateParentOrder(ctx, order); err != nil {
		return err
	}
	subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, orderNo)
	for _, so := range subOrders {
		so.Status = model.OrderStatusCompleted
		s.repo.UpdateSubOrder(ctx, &so)
	}
	return nil
}

// ApplyRefund 申请退款
func (s *OrderService) ApplyRefund(ctx context.Context, buyerID uint64, req *model.ApplyRefundRequest) error {
	subOrder, err := s.repo.FindSubOrderBySubOrderNo(ctx, req.SubOrderNo)
	if err != nil {
		return ErrOrderNotFound
	}
	if subOrder.BuyerID != buyerID {
		return ErrNotAuthorized
	}
	if subOrder.Status != model.OrderStatusPaid && subOrder.Status != model.OrderStatusDelivered {
		return ErrInvalidStatus
	}
	parentOrder, _ := s.repo.FindParentOrderByOrderNo(ctx, subOrder.ParentOrderNo)
	if parentOrder != nil && parentOrder.Status != model.OrderStatusDispute {
		parentOrder.Status = model.OrderStatusDispute
		s.repo.UpdateParentOrder(ctx, parentOrder)
	}
	subOrder.Status = model.OrderStatusDispute
	s.repo.UpdateSubOrder(ctx, subOrder)
	s.repo.CreateOrderActionLog(ctx, &model.OrderActionLog{
		OrderNo: subOrder.ParentOrderNo, Action: "apply_refund",
		Operator: fmt.Sprintf("%d", buyerID),
		Note: fmt.Sprintf("申请退款 type=%d reason=%s", req.Type, req.Reason),
	})
	return nil
}

// ============ 订单查询 ============

func (s *OrderService) GetOrderDetail(ctx context.Context, buyerID uint64, orderNo string) (*model.OrderResponse, error) {
	parentOrder, err := s.repo.FindParentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if parentOrder.BuyerID != buyerID {
		return nil, ErrNotAuthorized
	}
	subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, orderNo)
	addr, _ := s.repo.FindOrderAddressByOrderNo(ctx, orderNo)
	var subOrderInfos []model.SubOrderInfo
	for _, so := range subOrders {
		items, _ := s.repo.FindOrderItemsBySubOrderNo(ctx, so.SubOrderNo)
		subOrderInfos = append(subOrderInfos, model.SubOrderInfo{
			SubOrder: &so,
			Items:    items,
			ShopName: "",
		})
	}
	return &model.OrderResponse{
		ParentOrder: parentOrder,
		SubOrders:   subOrderInfos,
		Address:     addr,
	}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, buyerID uint64, q *model.OrderListQuery) (*model.OrderListResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	var status *uint8
	if q.Status > 0 {
		status = &q.Status
	}
	orders, total, err := s.repo.ListParentOrders(ctx, buyerID, status, q.Page, q.PageSize)
	if err != nil {
		return nil, err
	}
	var items []model.OrderListItem
	for _, order := range orders {
		subOrders, _ := s.repo.FindSubOrdersByParentOrderNo(ctx, order.OrderNo)
		itemCount := 0
		for _, so := range subOrders {
			orderItems, _ := s.repo.FindOrderItemsBySubOrderNo(ctx, so.SubOrderNo)
			itemCount += len(orderItems)
		}
		items = append(items, model.OrderListItem{
			OrderNo:     order.OrderNo,
			Status:      order.Status,
			StatusText:  model.OrderStatusText[order.Status],
			TotalAmount: order.TotalAmount,
			PayAmount:   order.PayAmount,
			ItemCount:   itemCount,
			CreatedAt:   order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &model.OrderListResponse{
		Total: total,
		Page:  q.Page,
		Size:  q.PageSize,
		Items: items,
	}, nil
}
