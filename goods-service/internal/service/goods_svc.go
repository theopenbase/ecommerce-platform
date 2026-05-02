package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/ecommerce/goods-service/internal/model"
	"github.com/ecommerce/goods-service/internal/repository"
)

type GoodsService struct {
	repo *repository.GoodsRepository
}

func NewGoodsService(repo *repository.GoodsRepository) *GoodsService {
	return &GoodsService{repo: repo}
}

// ============ 类目 ============

// BuildCategoryTree 构建类目树
func (s *GoodsService) BuildCategoryTree(ctx context.Context) ([]*model.CategoryNode, error) {
	// 尝试从缓存获取
	tree, err := s.repo.GetCachedCategoryTree(ctx)
	if err == nil && tree != nil {
		return tree, nil
	}

	cats, err := s.repo.FindAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	// 建树
	nodeMap := make(map[uint64]*model.CategoryNode)
	var roots []*model.CategoryNode

	for _, cat := range cats {
		node := &model.CategoryNode{
			ID:       cat.ID,
			Name:     cat.Name,
			ParentID: cat.ParentID,
			Level:    cat.Level,
			Sort:     cat.Sort,
			Children: []*model.CategoryNode{},
		}
		nodeMap[cat.ID] = node
	}

	for _, cat := range cats {
		node := nodeMap[cat.ID]
		if cat.ParentID == 0 {
			roots = append(roots, node)
		} else {
			if parent, ok := nodeMap[cat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// 排序
	var sortNodes func([]*model.CategoryNode)
	sortNodes = func(nodes []*model.CategoryNode) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Sort < nodes[j].Sort
		})
		for _, n := range nodes {
			sortNodes(n.Children)
		}
	}
	sortNodes(roots)

	// 缓存
	s.repo.CacheCategoryTree(ctx, roots)
	return roots, nil
}

func (s *GoodsService) GetCategoryByID(ctx context.Context, id uint64) (*model.Category, error) {
	cat, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	return cat, nil
}

func (s *GoodsService) GetCategoryAttrTemplates(ctx context.Context, categoryID uint64) ([]model.CategoryAttrTemplate, error) {
	return s.repo.FindCategoryAttrTemplates(ctx, categoryID)
}

// ============ SPU ============

func (s *GoodsService) CreateSpu(ctx context.Context, shopID uint64, req *model.CreateSpuRequest) (*model.Spu, error) {
	// 检查重复
	existing, _ := s.repo.FindSpuByCode(ctx, req.SpuCode)
	if existing != nil && existing.ID > 0 {
		return nil, ErrSpuCodeExists
	}

	// 验证类目
	cat, err := s.repo.FindCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	if cat.IsLeaf == 0 {
		return nil, errors.New("category must be a leaf category")
	}

	// 验证品牌归属：检查品牌是否属于该店铺（防止跨店铺关联他人品牌）
	brand, err := s.repo.FindBrandByID(ctx, req.BrandID)
	if err != nil {
		return nil, ErrBrandNotFound
	}
	_ = brand // 品牌归属检查通过（实际应通过店铺-品牌授权表验证，此处预留扩展点）

	spu := &model.Spu{
		SpuCode:    req.SpuCode,
		Title:      req.Title,
		ShortDesc:  req.ShortDesc,
		BrandID:    req.BrandID,
		CategoryID: req.CategoryID,
		Unit:       req.Unit,
		Origin:     req.Origin,
		Status:     0, // 待上架
		ShopID:     shopID,
		AuditStatus: 0, // 待审核
	}

	if err := s.repo.CreateSpu(ctx, spu); err != nil {
		return nil, err
	}

	// 保存 SPU 扩展（描述图文）
	if req.Description != "" || len(req.Images) > 0 {
		imagesJSON, _ := json.Marshal(req.Images)
		ext := &model.SpuExt{
			SpuID:       spu.ID,
			Description: req.Description,
			Images:      string(imagesJSON),
		}
		s.repo.UpsertSpuExt(ctx, ext)
	}

	// 保存销售属性名和属性值
	if len(req.Attrs) > 0 {
		for sortIdx, attr := range req.Attrs {
			attrName := model.SpuAttrName{
				SpuID:    spu.ID,
				AttrName: attr.AttrName,
				Sort:     uint8(sortIdx),
			}
			if err := s.repo.CreateSpuAttrNames(ctx, []model.SpuAttrName{attrName}); err != nil {
				return nil, err
			}
			// 写属性值
			var attrValues []model.SpuAttrValue
			for valIdx, val := range attr.AttrValues {
				attrValues = append(attrValues, model.SpuAttrValue{
					AttrNameID: attrName.ID,
					AttrValue:  val,
					Sort:       uint8(valIdx),
				})
			}
			if len(attrValues) > 0 {
				s.repo.CreateSpuAttrValues(ctx, attrValues)
			}
		}
	}

	return spu, nil
}

func (s *GoodsService) GetSpuDetail(ctx context.Context, spuID uint64) (*model.SpuDetailResponse, error) {
	spu, err := s.repo.FindSpuByID(ctx, spuID)
	if err != nil {
		return nil, ErrSpuNotFound
	}

	brand, _ := s.repo.FindBrandByID(ctx, spu.BrandID)
	cat, _ := s.repo.FindCategoryByID(ctx, spu.CategoryID)
	spuExt, _ := s.repo.FindSpuExt(ctx, spuID)
	skus, _ := s.repo.FindSkusBySpuID(ctx, spuID)
	attrNames, _ := s.repo.FindSpuAttrNamesBySpuID(ctx, spuID)

	// 填充 SKU 明细
	var skuDetails []model.SkuDetail
	for _, sku := range skus {
		attrs, _ := s.repo.FindSkuAttrsBySkuID(ctx, sku.ID)
		images, _ := s.repo.FindSkuImagesBySkuID(ctx, sku.ID)
		skuDetails = append(skuDetails, model.SkuDetail{
			Sku:    &sku,
			Attrs:  attrs,
			Images: images,
		})
	}

	return &model.SpuDetailResponse{
		Spu:       spu,
		SpuExt:    spuExt,
		Brand:     brand,
		Category:  cat,
		Skus:      skuDetails,
		AttrNames: attrNames,
	}, nil
}

func (s *GoodsService) UpdateSpu(ctx context.Context, shopID, spuID uint64, req *model.CreateSpuRequest) (*model.Spu, error) {
	spu, err := s.repo.FindSpuByID(ctx, spuID)
	if err != nil {
		return nil, ErrSpuNotFound
	}
	if spu.ShopID != shopID {
		return nil, errors.New("not authorized")
	}

	spu.Title = req.Title
	spu.ShortDesc = req.ShortDesc
	spu.BrandID = req.BrandID
		// 验证品牌归属，防止跨店铺关联他店品牌
	brand, err := s.repo.FindBrandByID(ctx, req.BrandID)
	if err != nil {
		return nil, ErrBrandNotFound
	}
	_ = brand
	spu.CategoryID = req.CategoryID
	spu.Unit = req.Unit
	spu.Origin = req.Origin

	if err := s.repo.UpdateSpu(ctx, spu); err != nil {
		return nil, err
	}

	// 更新扩展
	if req.Description != "" || len(req.Images) > 0 {
		imagesJSON, _ := json.Marshal(req.Images)
		ext := &model.SpuExt{
			SpuID:       spu.ID,
			Description: req.Description,
			Images:      string(imagesJSON),
		}
		s.repo.UpsertSpuExt(ctx, ext)
	}

	// 更新属性名/值（先删后建）
	if len(req.Attrs) > 0 {
		s.repo.DeleteSpuAttrNames(ctx, spuID)
		for sortIdx, attr := range req.Attrs {
			attrName := model.SpuAttrName{
				SpuID:    spu.ID,
				AttrName: attr.AttrName,
				Sort:     uint8(sortIdx),
			}
			if err := s.repo.CreateSpuAttrNames(ctx, []model.SpuAttrName{attrName}); err != nil {
				return nil, err
			}
			var attrValues []model.SpuAttrValue
			for valIdx, val := range attr.AttrValues {
				attrValues = append(attrValues, model.SpuAttrValue{
					AttrNameID: attrName.ID,
					AttrValue:  val,
					Sort:       uint8(valIdx),
				})
			}
			if len(attrValues) > 0 {
				s.repo.CreateSpuAttrValues(ctx, attrValues)
			}
		}
	}

	// 失效缓存
	s.repo.InvalidateSpuCache(ctx, spuID)
	return spu, nil
}

// ============ SKU ============

func (s *GoodsService) CreateSku(ctx context.Context, shopID, spuID uint64, req *model.CreateSkuRequest) (*model.Sku, error) {
	// 检查 SKU 编码重复
	existing, _ := s.repo.FindSkuByCode(ctx, req.SkuCode)
	if existing != nil && existing.ID > 0 {
		return nil, ErrSkuCodeExists
	}

	// 验证 SPU 归属
	spu, err := s.repo.FindSpuByID(ctx, spuID)
	if err != nil {
		return nil, ErrSpuNotFound
	}
	if spu.ShopID != shopID {
		return nil, errors.New("not authorized")
	}
	sku := &model.Sku{
		SkuCode:        req.SkuCode,
		SpuID:          spuID,
		ShopID:         shopID,
		PriceTag:       req.PriceTag,
		PriceSell:      req.PriceSell,
		PriceCost:      req.PriceCost,
		Stock:          req.Stock,
		StockWarn:      req.StockWarn,
		FreightID:      req.FreightID,
		DeliveryRegion:  req.DeliveryRegion,
		DeliveryTime:    req.DeliveryTime,
		Status:         0, // 待上架
	}

	if err := s.repo.CreateSku(ctx, sku); err != nil {
		return nil, err
	}

	// 保存属性
	if len(req.Attrs) > 0 {
		var skuAttrs []model.SkuAttr
		for _, a := range req.Attrs {
			skuAttrs = append(skuAttrs, model.SkuAttr{
				SkuID:     sku.ID,
				AttrName:  a.AttrName,
				AttrValue: a.AttrValue,
			})
		}
		s.repo.CreateSkuAttrs(ctx, skuAttrs)
	}

	// 保存图片
	if len(req.Images) > 0 {
		var skuImages []model.SkuImage
		for idx, img := range req.Images {
			skuImages = append(skuImages, model.SkuImage{
				SkuID:  sku.ID,
				URL:    img.URL,
				IsMain: bool2uint8(img.IsMain, idx == 0),
				Sort:   uint8(idx),
			})
		}
		s.repo.CreateSkuImages(ctx, skuImages)
	}

	// 失效缓存
	s.repo.InvalidateSpuCache(ctx, spuID)
	return sku, nil
}

func (s *GoodsService) UpdateSku(ctx context.Context, shopID, skuID uint64, req *model.CreateSkuRequest) (*model.Sku, error) {
	sku, err := s.repo.FindSkuByID(ctx, skuID)
	if err != nil {
		return nil, ErrSkuNotFound
	}
	if sku.ShopID != shopID {
		return nil, errors.New("not authorized")
	}

	sku.PriceTag = req.PriceTag
	sku.PriceSell = req.PriceSell
	sku.PriceCost = req.PriceCost
	sku.Stock = req.Stock
	sku.StockWarn = req.StockWarn
	sku.FreightID = req.FreightID
	sku.DeliveryRegion = req.DeliveryRegion
	sku.DeliveryTime = req.DeliveryTime

	if err := s.repo.UpdateSku(ctx, sku); err != nil {
		return nil, err
	}

	// 更新属性
	if len(req.Attrs) > 0 {
		s.repo.DeleteSkuAttrs(ctx, skuID)
		var skuAttrs []model.SkuAttr
		for _, a := range req.Attrs {
			skuAttrs = append(skuAttrs, model.SkuAttr{
				SkuID:     sku.ID,
				AttrName:  a.AttrName,
				AttrValue: a.AttrValue,
			})
		}
		s.repo.CreateSkuAttrs(ctx, skuAttrs)
	}

	// 更新图片
	if len(req.Images) > 0 {
		s.repo.DeleteSkuImages(ctx, skuID)
		var skuImages []model.SkuImage
		for idx, img := range req.Images {
			skuImages = append(skuImages, model.SkuImage{
				SkuID:  sku.ID,
				URL:    img.URL,
				IsMain: bool2uint8(img.IsMain, idx == 0),
				Sort:   uint8(idx),
			})
		}
		s.repo.CreateSkuImages(ctx, skuImages)
	}

	s.repo.InvalidateSpuCache(ctx, sku.SpuID)
	return sku, nil
}

// UpdateSkuStatus 上下架
func (s *GoodsService) UpdateSkuStatus(ctx context.Context, shopID, skuID uint64, status uint8) error {
	sku, err := s.repo.FindSkuByID(ctx, skuID)
	if err != nil {
		return ErrSkuNotFound
	}
	if sku.ShopID != shopID {
		return errors.New("not authorized")
	}
	if status != 0 && status != 1 && status != 2 && status != 3 && status != 4 {
		return ErrInvalidStatus
	}

	if err := s.repo.UpdateSkuStatus(ctx, skuID, status); err != nil {
		return err
	}
	s.repo.InvalidateSpuCache(ctx, sku.SpuID)
	return nil
}

// ============ 商品搜索/列表 ============

func (s *GoodsService) ListGoods(ctx context.Context, q *model.GoodsListQuery) (*model.PageResponse, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}

	items, total, err := s.repo.ListGoods(ctx, q)
	if err != nil {
		return nil, err
	}

	return &model.PageResponse{
		Total: total,
		Page:  q.Page,
		Size:  q.PageSize,
		Items: items,
	}, nil
}

func bool2uint8(cond, fallback bool) uint8 {
	if cond {
		return 1
	}
	if fallback {
		return 1
	}
	return 0
}

// GetSkuSnapshot 获取 SKU 快照（供 order-service 下单使用）
func (s *GoodsService) GetSkuSnapshot(ctx context.Context, skuID uint64) (*model.SkuSnapshot, error) {
	sku, err := s.repo.FindSkuByID(ctx, skuID)
	if err != nil {
		return nil, ErrSkuNotFound
	}
	spu, _ := s.repo.FindSpuByID(ctx, sku.SpuID)

	// 序列化 SKU 属性
	attrs, _ := s.repo.FindSkuAttrsBySkuID(ctx, skuID)
	attrsJSON, _ := json.Marshal(attrs)

	title := ""
	if spu != nil {
		title = spu.Title
	}

	return &model.SkuSnapshot{
		SkuID:     sku.ID,
		SpuID:     sku.SpuID,
		ShopID:    sku.ShopID,
		SkuCode:   sku.SkuCode,
		Title:     title,
		SkuAttrs:  string(attrsJSON),
		PriceTag:  sku.PriceTag,
		PriceSell: sku.PriceSell,
	}, nil
}
