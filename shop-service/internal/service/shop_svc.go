package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ecommerce/shop-service/internal/model"
	"github.com/ecommerce/shop-service/internal/repository"
)

var (
	ErrShopNotFound     = errors.New("shop not found")
	ErrNotShopOwner    = errors.New("not shop owner")
	ErrQualificationNotFound = errors.New("qualification not found")
)

type ShopService struct {
	repo *repository.ShopRepository
}

func NewShopService(repo *repository.ShopRepository) *ShopService {
	return &ShopService{repo: repo}
}

func generateShopCode() string {
	return fmt.Sprintf("S%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%100000)
}

// ============ 商家入驻 ============

func (s *ShopService) RegisterShop(ctx context.Context, ownerID uint64, req *model.RegisterShopRequest) (*model.Shop, error) {
	// 检查是否已有店铺
	existing, _ := s.repo.FindShopByOwnerID(ctx, ownerID)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("shop already registered")
	}

	shop := &model.Shop{
		ShopCode:    generateShopCode(),
		Name:        req.Name,
		Type:       req.Type,
		OwnerID:    ownerID,
		Status:     model.ShopStatusPending, // 待审核
		Province:   req.Province,
		City:       req.City,
		Description: req.Description,
	}
	if err := s.repo.CreateShop(ctx, shop); err != nil {
		return nil, err
	}
	return shop, nil
}

func (s *ShopService) GetShop(ctx context.Context, shopID uint64) (*model.Shop, error) {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	return shop, nil
}

func (s *ShopService) GetShopByOwner(ctx context.Context, ownerID uint64) (*model.Shop, error) {
	return s.repo.FindShopByOwnerID(ctx, ownerID)
}

func (s *ShopService) UpdateShop(ctx context.Context, ownerID, shopID uint64, req *model.UpdateShopRequest) (*model.Shop, error) {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	if shop.OwnerID != ownerID {
		return nil, ErrNotShopOwner
	}
	if req.Name != "" {
		shop.Name = req.Name
	}
	if req.BannerURL != "" {
		shop.BannerURL = req.BannerURL
	}
	if req.Description != "" {
		shop.Description = req.Description
	}
	if err := s.repo.UpdateShop(ctx, shop); err != nil {
		return nil, err
	}
	return shop, nil
}

// ============ 资质审核 ============

func (s *ShopService) SubmitQualification(ctx context.Context, ownerID, shopID uint64, qualType, certNo, frontURL, backURL string) (*model.ShopQualification, error) {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	if shop.OwnerID != ownerID {
		return nil, ErrNotShopOwner
	}
	q := &model.ShopQualification{
		ShopID:   shopID,
		QualType: qualType,
		CertNo:   certNo,
		FrontURL: frontURL,
		BackURL:  backURL,
		Status:   0,
	}
	if err := s.repo.CreateQualification(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *ShopService) AuditQualification(ctx context.Context, shopID, auditorID uint64, qualID uint64, approved bool) error {
	var qs []model.ShopQualification
	all, _ := s.repo.ListQualifications(ctx, shopID)
	for _, q := range all {
		if q.ID == qualID {
			qs = append(qs, q)
			break
		}
	}
	if len(qs) == 0 {
		return ErrQualificationNotFound
	}
	q := &qs[0]
	q.Status = 1
	if !approved {
		q.Status = 2
	}
	now := time.Now()
	q.AuditedAt = &now
	q.AuditorID = auditorID
	return s.repo.UpdateQualification(ctx, q)
}

func (s *ShopService) ApproveShop(ctx context.Context, shopID uint64) error {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return ErrShopNotFound
	}
	shop.Status = model.ShopStatusActive
	return s.repo.UpdateShop(ctx, shop)
}

// ============ DSR绩效 ============

func (s *ShopService) UpdateDSR(ctx context.Context, shopID uint64, productScore, serviceScore, logisticsScore float64) error {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return ErrShopNotFound
	}
	shop.DSRProduct = productScore
	shop.DSRService = serviceScore
	shop.DSRLogistics = logisticsScore
	return s.repo.UpdateShop(ctx, shop)
}

func (s *ShopService) GetShopPerformance(ctx context.Context, shopID uint64) (map[string]interface{}, error) {
	shop, err := s.repo.FindShopByID(ctx, shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	overall := (shop.DSRProduct + shop.DSRService + shop.DSRLogistics) / 3.0
	return map[string]interface{}{
		"shop_id":       shop.ID,
		"shop_name":     shop.Name,
		"dsr_product":   shop.DSRProduct,
		"dsr_service":   shop.DSRService,
		"dsr_logistics": shop.DSRLogistics,
		"dsr_overall":   overall,
		"follower_count": shop.FollowerCount,
	}, nil
}

// ============ 运费模板 ============

func (s *ShopService) CreateFreightTemplate(ctx context.Context, shopID uint64, req *model.CreateFreightTemplateRequest) (*model.FreightTemplate, error) {
	t := &model.FreightTemplate{
		ShopID:           shopID,
		Name:             req.Name,
		Type:            req.Type,
		IsFreeThreshold: req.IsFreeThreshold,
		FreeAmount:      req.FreeAmount,
		FreeNum:         req.FreeNum,
	}
	if err := s.repo.CreateFreightTemplate(ctx, t); err != nil {
		return nil, err
	}
	if len(req.Rules) > 0 {
		var rules []model.FreightRule
		for _, r := range req.Rules {
			rules = append(rules, model.FreightRule{
				TemplateID:    t.ID,
				ProvinceCodes: r.ProvinceCodes,
				FirstAmount:  r.FirstAmount,
				FirstNum:     r.FirstNum,
				AddAmount:    r.AddAmount,
				AddNum:       r.AddNum,
			})
		}
		s.repo.CreateFreightRules(ctx, rules)
	}
	return t, nil
}

func (s *ShopService) GetFreightTemplates(ctx context.Context, shopID uint64) ([]model.FreightTemplate, error) {
	return s.repo.FindFreightTemplatesByShopID(ctx, shopID)
}

func (s *ShopService) GetFreightTemplateDetail(ctx context.Context, shopID, templateID uint64) (*model.FreightTemplateResponse, error) {
	t, err := s.repo.FindFreightTemplateByID(ctx, templateID)
	if err != nil {
		return nil, errors.New("template not found")
	}
	if t.ShopID != shopID {
		return nil, ErrNotShopOwner
	}
	rules, _ := s.repo.FindFreightRulesByTemplateID(ctx, templateID)
	return &model.FreightTemplateResponse{Template: t, Rules: rules}, nil
}

func (s *ShopService) DeleteFreightTemplate(ctx context.Context, shopID, templateID uint64) error {
	return s.repo.DeleteFreightTemplate(ctx, templateID, shopID)
}

// ============ 店铺装修 ============

func (s *ShopService) SaveDecoration(ctx context.Context, shopID uint64, layout string) error {
	d := &model.ShopDecoration{
		ShopID: shopID,
		Layout: layout,
	}
	return s.repo.UpsertDecoration(ctx, d)
}

func (s *ShopService) GetDecoration(ctx context.Context, shopID uint64) (*model.ShopDecoration, error) {
	return s.repo.FindDecorationByShopID(ctx, shopID)
}

// ============ 保证金 ============

func (s *ShopService) FreezeDeposit(ctx context.Context, shopID uint64, amount float64) error {
	d := &model.ShopDeposit{
		ShopID: shopID,
		Amount: amount,
		Status: 0, // 冻结
	}
	now := time.Now()
	d.FreezeTime = &now
	return s.repo.CreateDeposit(ctx, d)
}
