-- 用户模块数据库初始化脚本
-- 数据库：ecommerce_user

CREATE DATABASE IF NOT EXISTS ecommerce_user DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE ecommerce_user;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    mobile          VARCHAR(20) NOT NULL UNIQUE COMMENT '手机号',
    nickname        VARCHAR(32) NOT NULL COMMENT '昵称',
    avatar_url      VARCHAR(512) DEFAULT '' COMMENT '头像URL',
    gender          TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0-未知 1-男 2-女',
    status          TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1-正常 0-封禁',
    password        VARCHAR(128) DEFAULT '' COMMENT '登录密码（bcrypt）',
    last_login      DATETIME DEFAULT NULL COMMENT '最后登录时间',
    last_ip         VARCHAR(45) DEFAULT '' COMMENT '最后登录IP',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_mobile (mobile),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 登录日志表
CREATE TABLE IF NOT EXISTS login_logs (
    id           BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id      BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    ip           VARCHAR(45) NOT NULL COMMENT '登录IP',
    device       VARCHAR(128) DEFAULT '' COMMENT '设备信息',
    status       TINYINT NOT NULL COMMENT '1-成功 0-失败',
    fail_reason  VARCHAR(128) DEFAULT '' COMMENT '失败原因',
    login_time   DATETIME NOT NULL COMMENT '登录时间',
    INDEX idx_user (user_id),
    INDEX idx_login_time (login_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';

-- 会员等级表
CREATE TABLE IF NOT EXISTS member_levels (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    level      TINYINT UNSIGNED NOT NULL UNIQUE COMMENT '等级',
    name       VARCHAR(16) NOT NULL COMMENT '等级名称',
    threshold  DECIMAL(12,2) NOT NULL COMMENT '累计消费门槛',
    rights     JSON COMMENT '权益列表',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员等级表';

-- 初始会员等级数据
INSERT INTO member_levels (level, name, threshold, rights) VALUES
(0, '注册会员', 0.00, '[]'),
(1, '铜牌会员', 500.00, '["部分商品9.5折","每月1张免邮券"]'),
(2, '银牌会员', 3000.00, '["部分商品9折","每月3张免邮券","生日礼包"]'),
(3, '金牌会员', 10000.00, '["部分商品8.5折","每月5张免邮券","专属客服","优先发货"]'),
(4, '黑金会员', 50000.00, '["部分商品8折","全年包邮","专属顾问","生日礼包","优先发货"]');

-- 会员账户表
CREATE TABLE IF NOT EXISTS members (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id        BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '用户ID',
    level          TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当前等级',
    total_spend    DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '累计实付金额（不含退款）',
    growth_value   BIGINT NOT NULL DEFAULT 0 COMMENT '成长值',
    grace_end_date DATE DEFAULT NULL COMMENT '降级风险提示截止日',
    upgraded_at    DATETIME DEFAULT NULL COMMENT '最近升级时间',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_level (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员账户表';

-- 积分账户表
CREATE TABLE IF NOT EXISTS points_accounts (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '用户ID',
    balance     BIGINT NOT NULL DEFAULT 0 COMMENT '积分余额',
    total_earned BIGINT NOT NULL DEFAULT 0 COMMENT '累计获取',
    total_spent BIGINT NOT NULL DEFAULT 0 COMMENT '累计消耗',
    expire_date DATE DEFAULT NULL COMMENT '即将过期日期',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分账户表';

-- 积分流水表
CREATE TABLE IF NOT EXISTS points_logs (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    order_no        VARCHAR(32) DEFAULT '' COMMENT '关联订单号',
    type            VARCHAR(16) NOT NULL COMMENT 'order/sign/eval/redeem/expire/refund',
    points          BIGINT NOT NULL COMMENT '正数获取，负数消耗',
    balance_before  BIGINT NOT NULL COMMENT '变动前余额',
    balance_after   BIGINT NOT NULL COMMENT '变动后余额',
    expire_date     DATE DEFAULT NULL COMMENT '本次积分过期日',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_order (order_no),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分流水表';

-- 收货地址表
CREATE TABLE IF NOT EXISTS receiver_addresses (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id        BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    receiver       VARCHAR(32) NOT NULL COMMENT '收货人',
    mobile         VARCHAR(20) NOT NULL COMMENT '手机号',
    province_code  VARCHAR(10) NOT NULL COMMENT '省代码',
    city_code      VARCHAR(10) NOT NULL COMMENT '市代码',
    district_code  VARCHAR(10) NOT NULL COMMENT '区代码',
    province_name  VARCHAR(32) NOT NULL COMMENT '省名称',
    city_name      VARCHAR(32) NOT NULL COMMENT '市名称',
    district_name  VARCHAR(32) NOT NULL COMMENT '区名称',
    detail         VARCHAR(256) NOT NULL COMMENT '详细地址',
    tag            VARCHAR(16) DEFAULT '' COMMENT '标签：家/公司/父母',
    is_default     TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认地址',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_is_default (user_id, is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收货地址表';
