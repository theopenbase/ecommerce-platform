-- 支付模块数据库
CREATE DATABASE IF NOT EXISTS ecommerce_payment DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce_payment;

CREATE TABLE IF NOT EXISTS payment_transactions (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    trans_no        VARCHAR(64) NOT NULL UNIQUE COMMENT '平台交易流水号',
    order_no       VARCHAR(32) NOT NULL COMMENT '关联订单号',
    sub_order_no   VARCHAR(32) DEFAULT '' COMMENT '子订单号',
    buyer_id       BIGINT UNSIGNED NOT NULL COMMENT '买家ID',
    channel        VARCHAR(16) NOT NULL COMMENT 'alipay/wechat/bank/balance',
    channel_trans_no VARCHAR(128) DEFAULT '' COMMENT '渠道方流水号',
    amount         DECIMAL(12,2) NOT NULL COMMENT '支付金额',
    status         TINYINT NOT NULL COMMENT '0-待支付 1-支付中 2-成功 3-失败 4-已退款',
    pay_time       DATETIME DEFAULT NULL COMMENT '支付成功时间',
    expire_time    DATETIME NOT NULL COMMENT '支付过期时间',
    callback_raw   TEXT COMMENT '回调原始报文',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_order (order_no),
    INDEX idx_buyer (buyer_id),
    INDEX idx_channel (channel),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS accounts (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id        BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    user_type      TINYINT NOT NULL COMMENT '1-买家 2-商家 3-平台',
    balance        DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '账户余额',
    freeze_balance DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '冻结金额',
    password_pay   VARCHAR(64) DEFAULT '' COMMENT '支付密码',
    status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 0-冻结',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_type (user_id, user_type),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS account_logs (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    account_id     BIGINT UNSIGNED NOT NULL COMMENT '账户ID',
    trans_no       VARCHAR(64) NOT NULL COMMENT '关联支付流水号',
    order_no       VARCHAR(32) DEFAULT '' COMMENT '关联订单号',
    type           VARCHAR(16) NOT NULL COMMENT 'recharge/pay/refund/settle/freeze/unfreeze',
    amount         DECIMAL(12,2) NOT NULL COMMENT '正数入账，负数出账',
    balance_before  DECIMAL(14,2) NOT NULL COMMENT '变动前余额',
    balance_after   DECIMAL(14,2) NOT NULL COMMENT '变动后余额',
    note           VARCHAR(256) DEFAULT '' COMMENT '备注',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_account (account_id),
    INDEX idx_order (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refund_records (
    id           BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    refund_no    VARCHAR(32) NOT NULL UNIQUE COMMENT '退款单号',
    trans_no     VARCHAR(64) NOT NULL COMMENT '原交易流水号',
    order_no     VARCHAR(32) NOT NULL COMMENT '关联订单号',
    buyer_id     BIGINT UNSIGNED NOT NULL COMMENT '买家ID',
    amount       DECIMAL(12,2) NOT NULL COMMENT '退款金额',
    status       TINYINT NOT NULL DEFAULT 0 COMMENT '0-处理中 1-成功 2-失败',
    reason       VARCHAR(256) DEFAULT '' COMMENT '退款原因',
    process_time DATETIME DEFAULT NULL COMMENT '处理时间',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_trans (trans_no),
    INDEX idx_buyer (buyer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
