package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/ecommerce/shop-service/internal/model"
)

type ShopRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewShopRepository(db *gorm.DB, cache *redis.Client) *ShopRepository {
	return &ShopRepository{db: db, cache: cache}
}

func (r *ShopRepository) CreateShop(ctx context.Context, s *model.Shop) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *ShopRepository) FindShopByID(ctx context.Context, id uint64) (*model.Shop, error) {
	var s model.Shop
	err := r.db.WithContext(ctx).First(&s, id).Error
	return &s, err
}

func (r *ShopRepository) FindShopByOwnerID(ctx context.Context, ownerID uint64) (*model.Shop, error) {
	var s model.Shop
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).First(&s).Error
	return &s, err
}

func (r *ShopRepository) UpdateShop(ctx context.Context, s *model.Shop) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *ShopRepository) ListShops(ctx context.Context, status *uint8, page, size int) ([]model.Shop, int64, error) {
	var shops []model.Shop
	query := r.db.WithContext(ctx).Model(&model.Shop{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var total int64
	query.Count(&total)
	offset := (page - 1) * size
	err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&shops).Error
	return shops, total, err
}

func (r *ShopRepository) CreateQualification(ctx context.Context, q *model.ShopQualification) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *ShopRepository) ListQualifications(ctx context.Context, shopID uint64) ([]model.ShopQualification, error) {
	var qs []model.ShopQualification
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&qs).Error
	return qs, err
}

func (r *ShopRepository) UpdateQualification(ctx context.Context, q *model.ShopQualification) error {
	return r.db.WithContext(ctx).Save(q).Error
}

func (r *ShopRepository) CreateDeposit(ctx context.Context, d *model.ShopDeposit) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *ShopRepository) FindDepositByShopID(ctx context.Context, shopID uint64) (*model.ShopDeposit, error) {
	var d model.ShopDeposit
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&d).Error
	return &d, err
}

func (r *ShopRepository) UpdateDeposit(ctx context.Context, d *model.ShopDeposit) error {
	return r.db.WithContext(ctx).Save(d).Error
}

// Freight
func (r *ShopRepository) CreateFreightTemplate(ctx context.Context, t *model.FreightTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *ShopRepository) FindFreightTemplatesByShopID(ctx context.Context, shopID uint64) ([]model.FreightTemplate, error) {
	var ts []model.FreightTemplate
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&ts).Error
	return ts, err
}

func (r *ShopRepository) FindFreightTemplateByID(ctx context.Context, id uint64) (*model.FreightTemplate, error) {
	var t model.FreightTemplate
	err := r.db.WithContext(ctx).First(&t, id).Error
	return &t, err
}

func (r *ShopRepository) DeleteFreightTemplate(ctx context.Context, id, shopID uint64) error {
	return r.db.WithContext(ctx).Where("id = ? AND shop_id = ?", id, shopID).Delete(&model.FreightTemplate{}).Error
}

func (r *ShopRepository) CreateFreightRules(ctx context.Context, rules []model.FreightRule) error {
	if len(rules) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rules).Error
}

func (r *ShopRepository) FindFreightRulesByTemplateID(ctx context.Context, templateID uint64) ([]model.FreightRule, error) {
	var rules []model.FreightRule
	err := r.db.WithContext(ctx).Where("template_id = ?", templateID).Find(&rules).Error
	return rules, err
}

func (r *ShopRepository) DeleteFreightRulesByTemplateID(ctx context.Context, templateID uint64) error {
	return r.db.WithContext(ctx).Where("template_id = ?", templateID).Delete(&model.FreightRule{}).Error
}

// Decoration
func (r *ShopRepository) UpsertDecoration(ctx context.Context, d *model.ShopDecoration) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *ShopRepository) FindDecorationByShopID(ctx context.Context, shopID uint64) (*model.ShopDecoration, error) {
	var d model.ShopDecoration
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&d).Error
	return &d, err
}
