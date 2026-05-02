-- ============================================================
-- 会员等级预置数据
-- Level 0: 注册会员
-- Level 1: 铜牌会员
-- Level 2: 银牌会员
-- Level 3: 金牌会员
-- Level 4: 钻石会员
-- Level 5: 旗舰会员
-- ============================================================

INSERT INTO member_levels (id, name, level_code, discount_rate, points_multiplier, threshold_amount, created_at, updated_at) VALUES
(0, '注册会员',   'REGISTERED', 1.00, 1.0,    0,    NOW(), NOW()),
(1, '铜牌会员',   'BRONZE',     0.98, 1.2,  500,    NOW(), NOW()),
(2, '银牌会员',   'SILVER',     0.96, 1.5, 2000,    NOW(), NOW()),
(3, '金牌会员',   'GOLD',       0.94, 2.0, 5000,    NOW(), NOW()),
(4, '钻石会员',   'DIAMOND',    0.92, 2.5,20000,    NOW(), NOW()),
(5, '旗舰会员',   'FLAGSHIP',   0.90, 3.0,50000,    NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  discount_rate = VALUES(discount_rate),
  points_multiplier = VALUES(points_multiplier),
  threshold_amount = VALUES(threshold_amount);
