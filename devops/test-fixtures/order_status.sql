-- ============================================================
-- 订单状态枚举预置数据
-- ============================================================

INSERT INTO order_status_lookup (status_code, status_name, is_cancellable, is_refundable, created_at, updated_at) VALUES
(0, '待付款（Pending payment）',       1, 0, NOW(), NOW()),
(1, '已取消（Cancelled）',            0, 0, NOW(), NOW()),
(2, '已付款/待发货（Paid/Pending）',  0, 1, NOW(), NOW()),
(3, '已发货/待收货（Delivered）',      0, 1, NOW(), NOW()),
(4, '已收货/待评价（Received）',       0, 1, NOW(), NOW()),
(5, '已完成（Completed）',             0, 0, NOW(), NOW()),
(6, '维权中（Dispute）',               0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  status_name = VALUES(status_name),
  is_cancellable = VALUES(is_cancellable),
  is_refundable = VALUES(is_refundable);
