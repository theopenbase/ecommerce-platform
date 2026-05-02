package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/ecommerce/goods-service/internal/model"
)

type GoodsRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewGoodsRepository(db *gorm.DB, cache *redis.Client) *GoodsRepository {
	return &GoodsRepository{db: db, cache: cache}
}

// ============ 类目 ============

func (r *GoodsRepository) FindAllCategories(ctx context.Context) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.WithContext(ctx).Order("level asc, sort asc").Find(&cats).Error
	return cats, err
}

func (r *GoodsRepository) FindCategoryByID(ctx context.Context, id uint64) (*model.Category, error) {
	var cat model.Category
	err := r.db.WithContext(ctx).First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *GoodsRepository) FindCategoryByParentID(ctx context.Context, parentID uint64) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort asc").Find(&cats).Error
	return cats, err
}

func (r *GoodsRepository) CreateCategory(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *GoodsRepository) UpdateCategory(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

// ============ 品牌 ============

func (r *GoodsRepository) FindAllBrands(ctx context.Context) ([]model.Brand, error) {
	var brands []model.Brand
	err := r.db.WithContext(ctx).Where("status = 1").Order("name asc").Find(&brands).Error
	return brands, err
}

func (r *GoodsRepository) FindBrandByID(ctx context.Context, id uint64) (*model.Brand, error) {
	var brand model.Brand
	err := r.db.WithContext(ctx).First(&brand, id).Error
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// ============ SPU ============

func (r *GoodsRepository) CreateSpu(ctx context.Context, spu *model.Spu) error {
	return r.db.WithContext(ctx).Create(spu).Error
}

func (r *GoodsRepository) FindSpuByID(ctx context.Context, id uint64) (*model.Spu, error) {
	var spu model.Spu
	err := r.db.WithContext(ctx).First(&spu, id).Error
	if err != nil {
		return nil, err
	}
	return &spu, nil
}

func (r *GoodsRepository) FindSpuByCode(ctx context.Context, code string) (*model.Spu, error) {
	var spu model.Spu
	err := r.db.WithContext(ctx).Where("spu_code = ?", code).First(&spu).Error
	if err != nil {
		return nil, err
	}
	return &spu, nil
}

func (r *GoodsRepository) UpdateSpu(ctx context.Context, spu *model.Spu) error {
	return r.db.WithContext(ctx).Save(spu).Error
}

// ============ SPU 扩展 ============

func (r *GoodsRepository) UpsertSpuExt(ctx context.Context, ext *model.SpuExt) error {
	return r.db.WithContext(ctx).Save(ext).Error
}

func (r *GoodsRepository) FindSpuExt(ctx context.Context, spuID uint64) (*model.SpuExt, error) {
	var ext model.SpuExt
	err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).First(&ext).Error
	if err != nil {
		return nil, err
	}
	return &ext, nil
}

// ============ SKU ============

func (r *GoodsRepository) CreateSku(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *GoodsRepository) FindSkuByID(ctx context.Context, id uint64) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).First(&sku, id).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *GoodsRepository) FindSkuByCode(ctx context.Context, code string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Where("sku_code = ?", code).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *GoodsRepository) FindSkusBySpuID(ctx context.Context, spuID uint64) ([]model.Sku, error) {
	var skus []model.Sku
	err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).Find(&skus).Error
	return skus, err
}

func (r *GoodsRepository) UpdateSku(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *GoodsRepository) UpdateSkuStatus(ctx context.Context, id uint64, status uint8) error {
	return r.db.WithContext(ctx).Model(&model.Sku{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GoodsRepository) UpdateSpuStatus(ctx context.Context, id uint64, status uint8) error {
	return r.db.WithContext(ctx).Model(&model.Spu{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GoodsRepository) UpdateSpuAuditStatus(ctx context.Context, id uint64, auditStatus uint8) error {
	return r.db.WithContext(ctx).Model(&model.Spu{}).Where("id = ?", id).Update("audit_status", auditStatus).Error
}

// ============ SKU 属性 ============

func (r *GoodsRepository) CreateSkuAttrs(ctx context.Context, attrs []model.SkuAttr) error {
	if len(attrs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&attrs).Error
}

func (r *GoodsRepository) FindSkuAttrsBySkuID(ctx context.Context, skuID uint64) ([]model.SkuAttr, error) {
	var attrs []model.SkuAttr
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Find(&attrs).Error
	return attrs, err
}

func (r *GoodsRepository) DeleteSkuAttrs(ctx context.Context, skuID uint64) error {
	return r.db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&model.SkuAttr{}).Error
}

// ============ SKU 图片 ============

func (r *GoodsRepository) CreateSkuImages(ctx context.Context, images []model.SkuImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&images).Error
}

func (r *GoodsRepository) FindSkuImagesBySkuID(ctx context.Context, skuID uint64) ([]model.SkuImage, error) {
	var images []model.SkuImage
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).Order("is_main desc, sort asc").Find(&images).Error
	return images, err
}

func (r *GoodsRepository) DeleteSkuImages(ctx context.Context, skuID uint64) error {
	return r.db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&model.SkuImage{}).Error
}

// ============ SPU 销售属性名/值 ============

func (r *GoodsRepository) CreateSpuAttrNames(ctx context.Context, attrs []model.SpuAttrName) error {
	if len(attrs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&attrs).Error
}

func (r *GoodsRepository) FindSpuAttrNamesBySpuID(ctx context.Context, spuID uint64) ([]model.SpuAttrName, error) {
	var attrs []model.SpuAttrName
	err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).Order("sort asc").Find(&attrs).Error
	return attrs, err
}

func (r *GoodsRepository) CreateSpuAttrValues(ctx context.Context, values []model.SpuAttrValue) error {
	if len(values) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&values).Error
}

func (r *GoodsRepository) FindSpuAttrValuesByAttrNameIDs(ctx context.Context, nameIDs []uint64) ([]model.SpuAttrValue, error) {
	var values []model.SpuAttrValue
	err := r.db.WithContext(ctx).Where("attr_name_id IN ?", nameIDs).Find(&values).Error
	return values, err
}

func (r *GoodsRepository) DeleteSpuAttrNames(ctx context.Context, spuID uint64) error {
	return r.db.WithContext(ctx).Where("spu_id = ?", spuID).Delete(&model.SpuAttrName{}).Error
}

// ============ 类目属性模板 ============

func (r *GoodsRepository) FindCategoryAttrTemplates(ctx context.Context, categoryID uint64) ([]model.CategoryAttrTemplate, error) {
	var templates []model.CategoryAttrTemplate
	err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).Order("sort asc").Find(&templates).Error
	return templates, err
}

// ============ 商品列表（分页 + 过滤）============

func (r *GoodsRepository) ListGoods(ctx context.Context, q *model.GoodsListQuery) ([]model.GoodsListItem, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Spu{}).
		Select("spus.id as spu_id, spus.title, spus.short_desc, spus.brand_id, brands.name as brand_name, spus.category_id, MIN(skus.price_sell) as min_price, MAX(skus.price_sell) as max_price, sku_images.url as main_image, SUM(skus.stock) as stock, 0 as sales_count").
		Joins("LEFT JOIN brands ON spus.brand_id = brands.id").
		Joins("LEFT JOIN skus ON spus.id = skus.spu_id AND skus.is_main = 1 OR skus.id IN (SELECT id FROM skus GROUP BY spu_id)").
		Joins("LEFT JOIN (SELECT sku_id, url FROM sku_images WHERE is_main = 1) AS sku_images ON skus.id = sku_images.sku_id").
		Where("spus.status = ?", 1)

	if q.CategoryID > 0 {
		query = query.Where("spus.category_id = ?", q.CategoryID)
	}
	if q.BrandID > 0 {
		query = query.Where("spus.brand_id = ?", q.BrandID)
	}
	if q.Keyword != "" {
		query = query.Where("spus.title LIKE ? OR spus.short_desc LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	if q.PriceMin > 0 {
		query = query.Where("skus.price_sell >= ?", q.PriceMin)
	}
	if q.PriceMax > 0 {
		query = query.Where("skus.price_sell <= ?", q.PriceMax)
	}
	if q.ShopID > 0 {
		query = query.Where("spus.shop_id = ?", q.ShopID)
	}

	var total int64
	query.Count(&total)

	offset := (q.Page - 1) * q.PageSize
	if offset < 0 {
		offset = 0
	}

	var items []model.GoodsListItem
	err := query.Group("spus.id").Order("spus.created_at DESC").Offset(offset).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

// ============ 缓存 ============

func (r *GoodsRepository) CacheCategoryTree(ctx context.Context, tree []*model.CategoryNode) error {
	key := "goods:category:tree"
	data, _ := json.Marshal(tree)
	return r.cache.Set(ctx, key, data, 30*time.Minute).Err()
}

func (r *GoodsRepository) GetCachedCategoryTree(ctx context.Context) ([]*model.CategoryNode, error) {
	key := "goods:category:tree"
	data, err := r.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var tree []*model.CategoryNode
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

func (r *GoodsRepository) InvalidateCategoryTree(ctx context.Context) error {
	key := "goods:category:tree"
	return r.cache.Del(ctx, key).Err()
}

func (r *GoodsRepository) CacheSpu(ctx context.Context, spuID uint64, data []byte) error {
	key := fmt.Sprintf("goods:spu:%d", spuID)
	return r.cache.Set(ctx, key, data, 5*time.Minute).Err()
}

func (r *GoodsRepository) GetCachedSpu(ctx context.Context, spuID uint64) ([]byte, error) {
	key := fmt.Sprintf("goods:spu:%d", spuID)
	return r.cache.Get(ctx, key).Bytes()
}

func (r *GoodsRepository) InvalidateSpuCache(ctx context.Context, spuID uint64) error {
	key := fmt.Sprintf("goods:spu:%d", spuID)
	return r.cache.Del(ctx, key).Err()
}
