-- 订单模块数据库初始化脚本
-- 数据库：ecommerce_order

CREATE DATABASE IF NOT EXISTS ecommerce_order DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce_order;

-- 购物车表
CREATE TABLE IF NOT EXISTS carts (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    sku_id     BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    spu_id     BIGINT UNSIGNED NOT NULL COMMENT 'SPU ID',
    shop_id    BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    quantity   INT NOT NULL DEFAULT 1 COMMENT '购买数量',
    checked    TINYINT NOT NULL DEFAULT 1 COMMENT '1-选中 0-未选中',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_sku (user_id, sku_id),
    INDEX idx_user (user_id),
    INDEX idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';

-- 父订单表
CREATE TABLE IF NOT EXISTS parent_orders (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_no       VARCHAR(32) NOT NULL UNIQUE COMMENT '订单编号',
    buyer_id       BIGINT UNSIGNED NOT NULL COMMENT '买家ID',
    total_amount   DECIMAL(14,2) NOT NULL COMMENT '商品总额',
    freight_amount DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '运费',
    discount_amt   DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '优惠金额',
    pay_amount     DECIMAL(14,2) NOT NULL COMMENT '实付金额',
    status         TINYINT NOT NULL COMMENT '0-待付款 1-已取消 2-待发货 3-待收货 4-已收货 5-已完成 6-维权中',
    pay_time       DATETIME DEFAULT NULL COMMENT '付款时间',
    delivery_time  DATETIME DEFAULT NULL COMMENT '发货时间',
    receive_time   DATETIME DEFAULT NULL COMMENT '收货时间',
    finish_time    DATETIME DEFAULT NULL COMMENT '完成时间',
    cancel_time    DATETIME DEFAULT NULL COMMENT '取消时间',
    cancel_reason  VARCHAR(256) DEFAULT '' COMMENT '取消原因',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_buyer (buyer_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='父订单表';

-- 子订单表
CREATE TABLE IF NOT EXISTS sub_orders (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sub_order_no    VARCHAR(32) NOT NULL UNIQUE COMMENT '子订单号',
    parent_order_no VARCHAR(32) NOT NULL COMMENT '父订单号',
    buyer_id        BIGINT UNSIGNED NOT NULL COMMENT '买家ID',
    shop_id         BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    status          TINYINT NOT NULL COMMENT '同父订单状态语义',
    freight_amount  DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '子订单运费',
    discount_amt    DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '子订单优惠',
    pay_amount      DECIMAL(14,2) NOT NULL COMMENT '子订单实付',
    invoice_type    TINYINT DEFAULT 0 COMMENT '0-不开发票 1-普票 2-增票',
    invoice_title   VARCHAR(128) DEFAULT '' COMMENT '发票抬头',
    invoice_tax_no  VARCHAR(32) DEFAULT '' COMMENT '税号',
    logistics_no    VARCHAR(64) DEFAULT '' COMMENT '物流单号',
    logistics_co    VARCHAR(32) DEFAULT '' COMMENT '物流公司',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent (parent_order_no),
    INDEX idx_buyer (buyer_id),
    INDEX idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='子订单表';

-- 订单商品项表
CREATE TABLE IF NOT EXISTS order_items (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sub_order_no  VARCHAR(32) NOT NULL COMMENT '子订单号',
    sku_id         BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    spu_id         BIGINT UNSIGNED NOT NULL COMMENT 'SPU ID',
    sku_code       VARCHAR(64) NOT NULL COMMENT 'SKU编码快照',
    title          VARCHAR(80) NOT NULL COMMENT '商品标题快照',
    sku_attrs      VARCHAR(512) NOT NULL COMMENT 'SKU属性快照JSON',
    price_tag      DECIMAL(12,2) NOT NULL COMMENT '吊牌价快照',
    price_sell     DECIMAL(12,2) NOT NULL COMMENT '销售价快照',
    quantity       INT NOT NULL COMMENT '购买数量',
    discount_amt   DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '商品优惠分摊',
    item_total     DECIMAL(14,2) NOT NULL COMMENT '小计',
    refund_status  TINYINT NOT NULL DEFAULT 0 COMMENT '0-无退款 1-部分退款 2-全额退款',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sub_order (sub_order_no),
    INDEX idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品项表';

-- 订单收货地址快照
CREATE TABLE IF NOT EXISTS order_addresses (
    id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_no      VARCHAR(32) NOT NULL UNIQUE COMMENT '父订单号',
    receiver      VARCHAR(32) NOT NULL COMMENT '收货人',
    mobile        VARCHAR(20) NOT NULL COMMENT '手机号',
    province_code VARCHAR(10) NOT NULL COMMENT '省代码',
    city_code     VARCHAR(10) NOT NULL COMMENT '市代码',
    district_code VARCHAR(10) NOT NULL COMMENT '区代码',
    province_name VARCHAR(32) NOT NULL COMMENT '省名称',
    city_name     VARCHAR(32) NOT NULL COMMENT '市名称',
    district_name VARCHAR(32) NOT NULL COMMENT '区名称',
    detail        VARCHAR(256) NOT NULL COMMENT '详细地址',
    INDEX idx_order (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单收货地址快照表';

-- 订单操作日志
CREATE TABLE IF NOT EXISTS order_action_logs (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_no   VARCHAR(32) NOT NULL COMMENT '订单号',
    action     VARCHAR(32) NOT NULL COMMENT '操作类型',
    operator   VARCHAR(32) DEFAULT '' COMMENT '操作人',
    note       VARCHAR(256) DEFAULT '' COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order (order_no),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单操作日志表';
