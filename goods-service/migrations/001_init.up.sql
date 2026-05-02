-- 商品模块数据库初始化脚本
-- 数据库：ecommerce_goods

CREATE DATABASE IF NOT EXISTS ecommerce_goods DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce_goods;

-- 类目表（邻接表模型）
CREATE TABLE IF NOT EXISTS categories (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name       VARCHAR(32) NOT NULL COMMENT '类目名称',
    parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父类目ID，0为根',
    level      TINYINT UNSIGNED NOT NULL COMMENT '层级1/2/3/4',
    path       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '路径 /1/3/5/',
    sort       INT NOT NULL DEFAULT 0 COMMENT '排序',
    status     TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 0-禁用',
    is_leaf    TINYINT NOT NULL DEFAULT 0 COMMENT '是否叶子类目',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent (parent_id),
    INDEX idx_level (level),
    INDEX idx_path (path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='类目表';

-- 品牌表
CREATE TABLE IF NOT EXISTS brands (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name       VARCHAR(64) NOT NULL UNIQUE COMMENT '品牌名称',
    logo_url   VARCHAR(512) DEFAULT '' COMMENT '品牌Logo',
    status     TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 0-禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='品牌表';

-- SPU表（标准化产品单元）
CREATE TABLE IF NOT EXISTS spus (
    id           BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    spu_code     VARCHAR(32) NOT NULL UNIQUE COMMENT 'SPU编码',
    title        VARCHAR(80) NOT NULL COMMENT '商品标题',
    short_desc   VARCHAR(200) NOT NULL COMMENT '短描述',
    brand_id     BIGINT UNSIGNED NOT NULL COMMENT '品牌ID',
    category_id  BIGINT UNSIGNED NOT NULL COMMENT '三级类目ID',
    unit         VARCHAR(16) NOT NULL COMMENT '单位',
    origin       VARCHAR(64) DEFAULT '' COMMENT '产地',
    status       TINYINT NOT NULL DEFAULT 0 COMMENT '0-待上架 1-上架 2-下架 3-归档',
    shop_id      BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    audit_status TINYINT NOT NULL DEFAULT 0 COMMENT '0-待审核 1-通过 2-拒绝',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category_id),
    INDEX idx_brand (brand_id),
    INDEX idx_shop (shop_id),
    INDEX idx_status (status),
    INDEX idx_audit (audit_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SPU表';

-- SPU扩展表（图文描述）
CREATE TABLE IF NOT EXISTS spu_exts (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    spu_id      BIGINT UNSIGNED NOT NULL UNIQUE COMMENT 'SPU ID',
    description TEXT COMMENT '商品详情（富文本）',
    images      VARCHAR(2048) DEFAULT '' COMMENT '图文描述图片JSON数组',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_spu (spu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SPU扩展表';

-- SKU表（具体规格售卖单元）
CREATE TABLE IF NOT EXISTS skus (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sku_code        VARCHAR(64) NOT NULL UNIQUE COMMENT 'SKU编码（商家自定义）',
    spu_id          BIGINT UNSIGNED NOT NULL COMMENT '所属SPU ID',
    shop_id         BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    price_tag       DECIMAL(12,2) NOT NULL COMMENT '吊牌价（划线价）',
    price_sell      DECIMAL(12,2) NOT NULL COMMENT '销售价',
    price_cost      DECIMAL(12,2) DEFAULT NULL COMMENT '成本价（商家可见）',
    stock           INT NOT NULL DEFAULT 0 COMMENT '库存数量',
    stock_warn      INT DEFAULT 0 COMMENT '库存预警值',
    freight_id      BIGINT UNSIGNED NOT NULL COMMENT '运费模板ID',
    delivery_region VARCHAR(512) NOT NULL COMMENT '发货地（省市区）',
    delivery_time   TINYINT UNSIGNED NOT NULL COMMENT '发货时间（小时）',
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '0-待上架 1-上架 2-下架 3-售罄 4-归档',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_sku_code (sku_code),
    INDEX idx_spu (spu_id),
    INDEX idx_shop (shop_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SKU表';

-- SKU属性值表（具体规格组合，如颜色/尺码）
CREATE TABLE IF NOT EXISTS sku_attrs (
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sku_id     BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    attr_name  VARCHAR(32) NOT NULL COMMENT '属性名，如颜色/尺码',
    attr_value VARCHAR(128) NOT NULL COMMENT '属性值，如红色/L',
    INDEX idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SKU属性表';

-- SKU图片表
CREATE TABLE IF NOT EXISTS sku_images (
    id      BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    sku_id  BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    url     VARCHAR(512) NOT NULL COMMENT '图片OSS URL',
    is_main TINYINT NOT NULL DEFAULT 0 COMMENT '是否主图',
    sort    TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    INDEX idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SKU图片表';

-- 类目属性模板表（末级类目绑定属性名集合）
CREATE TABLE IF NOT EXISTS category_attr_templates (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    category_id BIGINT UNSIGNED NOT NULL COMMENT '末级类目ID',
    attr_name   VARCHAR(32) NOT NULL COMMENT '属性名',
    attr_type   VARCHAR(16) NOT NULL COMMENT '类型：text/number/boolean/multi',
    is_required TINYINT NOT NULL DEFAULT 0 COMMENT '是否必填',
    sort        INT NOT NULL DEFAULT 0 COMMENT '排序',
    INDEX idx_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='类目属性模板';

-- SPU销售属性名（跨SKU共享，如颜色/尺码）
CREATE TABLE IF NOT EXISTS spu_attr_names (
    id       BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    spu_id   BIGINT UNSIGNED NOT NULL COMMENT 'SPU ID',
    attr_name VARCHAR(32) NOT NULL COMMENT '属性名',
    sort     TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    INDEX idx_spu (spu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SPU销售属性名';

-- SPU销售属性值
CREATE TABLE IF NOT EXISTS spu_attr_values (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    attr_name_id BIGINT UNSIGNED NOT NULL COMMENT '属性名ID',
    attr_value  VARCHAR(128) NOT NULL COMMENT '属性值',
    sort        TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序',
    INDEX idx_attr_name (attr_name_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SPU销售属性值';
