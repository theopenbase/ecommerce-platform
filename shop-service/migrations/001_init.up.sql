-- 店铺模块数据库
CREATE DATABASE IF NOT EXISTS ecommerce_shop DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce_shop;

CREATE TABLE IF NOT EXISTS shops (
    id             BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    shop_code     VARCHAR(32) NOT NULL UNIQUE COMMENT '店铺编码',
    name          VARCHAR(64) NOT NULL COMMENT '店铺名称',
    type          TINYINT NOT NULL COMMENT '1-自营 2-旗舰 3-专卖 4-专营',
    owner_id     BIGINT UNSIGNED NOT NULL COMMENT '店主用户ID',
    status        TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审核 1-正常 2-封禁 3-清退',
    logo_url     VARCHAR(512) DEFAULT '' COMMENT '店铺Logo',
    banner_url   VARCHAR(512) DEFAULT '' COMMENT '店招Banner',
    description  VARCHAR(500) DEFAULT '' COMMENT '店铺简介',
    province     VARCHAR(32) DEFAULT '' COMMENT '省份',
    city         VARCHAR(32) DEFAULT '' COMMENT '城市',
    dsr_product  DECIMAL(3,2) DEFAULT 0 COMMENT '商品描述DSR',
    dsr_service  DECIMAL(3,2) DEFAULT 0 COMMENT '服务态度DSR',
    dsr_logistics DECIMAL(3,2) DEFAULT 0 COMMENT '物流速度DSR',
    follower_count BIGINT UNSIGNED DEFAULT 0 COMMENT '关注数',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_owner (owner_id),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shop_qualifications (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    shop_id    BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    qual_type  VARCHAR(32) NOT NULL COMMENT '资质类型',
    cert_no    VARCHAR(64) DEFAULT '' COMMENT '证书编号',
    front_url  VARCHAR(512) NOT NULL COMMENT '正面图片URL',
    back_url   VARCHAR(512) DEFAULT '' COMMENT '背面图片URL',
    expiry_date DATE DEFAULT NULL COMMENT '有效期',
    status     TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审 1-通过 2-拒绝',
    audited_at DATETIME DEFAULT NULL,
    auditor_id BIGINT UNSIGNED DEFAULT NULL,
    INDEX idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shop_deposits (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    shop_id    BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    amount     DECIMAL(12,2) NOT NULL COMMENT '保证金金额',
    status     TINYINT NOT NULL DEFAULT 0 COMMENT '0-冻结 1-可用 2-扣除中',
    freeze_time DATETIME DEFAULT NULL,
    INDEX idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS freight_templates (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    shop_id        BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    name           VARCHAR(32) NOT NULL COMMENT '模板名称',
    type           TINYINT NOT NULL COMMENT '1-按件 2-按重量 3-固定 4-地区计价',
    is_free_threshold TINYINT NOT NULL DEFAULT 0 COMMENT '是否满足条件包邮',
    free_amount    DECIMAL(12,2) DEFAULT 0 COMMENT '满X元包邮',
    free_num       INT DEFAULT 0 COMMENT '满X件包邮',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS freight_rules (
    id               BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    template_id     BIGINT UNSIGNED NOT NULL COMMENT '运费模板ID',
    province_codes  VARCHAR(256) NOT NULL COMMENT '省市区代码，逗号分隔',
    first_amount    DECIMAL(10,2) NOT NULL COMMENT '首费',
    first_num       INT NOT NULL DEFAULT 1 COMMENT '首件数量/首重kg',
    add_amount      DECIMAL(10,2) NOT NULL COMMENT '续费',
    add_num         INT NOT NULL DEFAULT 1 COMMENT '续件数量/续重kg',
    INDEX idx_template (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shop_decorations (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    shop_id   BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '店铺ID',
    layout    JSON NOT NULL COMMENT '装修布局配置',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
