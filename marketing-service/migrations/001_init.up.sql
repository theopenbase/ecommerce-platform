-- 营销模块数据库
CREATE DATABASE IF NOT EXISTS ecommerce_marketing DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce_marketing;

CREATE TABLE IF NOT EXISTS coupons (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    coupon_code    VARCHAR(32) NOT NULL UNIQUE COMMENT '券模板编码',
    name           VARCHAR(32) NOT NULL COMMENT '优惠券名称',
    type           TINYINT NOT NULL COMMENT '1-平台券 2-店铺券 3-商品券 4-内部券',
    face_value     DECIMAL(10,2) NOT NULL COMMENT '面额',
    threshold      DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '满减门槛',
    total_count    INT NOT NULL COMMENT '发放总量',
    remain_count   INT NOT NULL COMMENT '剩余数量',
    per_user_limit INT NOT NULL DEFAULT 1 COMMENT '每人限领',
    valid_type     TINYINT NOT NULL COMMENT '1-绝对时间 2-相对领取日',
    valid_start    DATETIME DEFAULT NULL,
    valid_end      DATETIME DEFAULT NULL,
    valid_days     INT DEFAULT NULL COMMENT '领取后N天有效',
    applicable_type TINYINT NOT NULL DEFAULT 0 COMMENT '0-全品 1-指定类目 2-指定商品 3-指定店铺',
    applicable_ids VARCHAR(512) DEFAULT '' COMMENT '关联ID列表',
    shop_id        BIGINT UNSIGNED DEFAULT NULL COMMENT '店铺券时关联店铺',
    status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-待发放 2-发放中 3-已领完 4-已过期',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_shop (shop_id),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_coupons (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    coupon_id  BIGINT UNSIGNED NOT NULL COMMENT '优惠券ID',
    coupon_code VARCHAR(32) NOT NULL COMMENT '用户领取到的券实例编码',
    status     TINYINT NOT NULL DEFAULT 0 COMMENT '0-未使用 1-已使用 2-已过期 3-已退款',
    received_at DATETIME NOT NULL COMMENT '领取时间',
    used_at    DATETIME DEFAULT NULL,
    used_order_no VARCHAR(32) DEFAULT '' COMMENT '使用的订单号',
    expire_date DATE NOT NULL COMMENT '过期日期',
    INDEX idx_user (user_id),
    INDEX idx_status (status),
    INDEX idx_expire (expire_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS promotions (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name       VARCHAR(64) NOT NULL COMMENT '活动名称',
    type       VARCHAR(16) NOT NULL COMMENT 'full_reduce/flash_seckill/group_buy/tier_price/pre_sell/discount',
    start_time DATETIME NOT NULL,
    end_time   DATETIME NOT NULL,
    status     TINYINT NOT NULL DEFAULT 0 COMMENT '0-未开始 1-进行中 2-已结束',
    rules      JSON NOT NULL COMMENT '促销规则',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_time (start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS promotion_skus (
    id           BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    promotion_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID',
    sku_id       BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    shop_id      BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    promo_price  DECIMAL(12,2) NOT NULL COMMENT '活动价格',
    stock_limit  INT NOT NULL COMMENT '活动库存上限',
    sold_count   INT NOT NULL DEFAULT 0 COMMENT '已售数量',
    INDEX idx_promotion (promotion_id),
    INDEX idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
