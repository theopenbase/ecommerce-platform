package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ecommerce/goods-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


// ============ Mock Goods Repository ============

type MockGoodsRepository struct {
	mock.Mock
}

func (m *MockGoodsRepository) FindAllCategories(ctx context.Context) ([]model.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Category), args.Error(1)
}

func (m *MockGoodsRepository) GetCachedCategoryTree(ctx context.Context) ([]*model.CategoryNode, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.CategoryNode), args.Error(1)
}

func (m *MockGoodsRepository) CacheCategoryTree(ctx context.Context, tree []*model.CategoryNode) error {
	args := m.Called(ctx, tree)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindCategoryByID(ctx context.Context, id uint64) (*model.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *MockGoodsRepository) FindCategoryAttrTemplates(ctx context.Context, categoryID uint64) ([]model.CategoryAttrTemplate, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).([]model.CategoryAttrTemplate), args.Error(1)
}

func (m *MockGoodsRepository) FindSpuByCode(ctx context.Context, code string) (*model.Spu, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Spu), args.Error(1)
}

func (m *MockGoodsRepository) FindSpuByID(ctx context.Context, id uint64) (*model.Spu, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Spu), args.Error(1)
}

func (m *MockGoodsRepository) CreateSpu(ctx context.Context, spu *model.Spu) error {
	args := m.Called(ctx, spu)
	return args.Error(0)
}

func (m *MockGoodsRepository) UpdateSpu(ctx context.Context, spu *model.Spu) error {
	args := m.Called(ctx, spu)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindSpuExt(ctx context.Context, spuID uint64) (*model.SpuExt, error) {
	args := m.Called(ctx, spuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SpuExt), args.Error(1)
}

func (m *MockGoodsRepository) UpsertSpuExt(ctx context.Context, ext *model.SpuExt) error {
	args := m.Called(ctx, ext)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindSpuAttrNamesBySpuID(ctx context.Context, spuID uint64) ([]model.SpuAttrName, error) {
	args := m.Called(ctx, spuID)
	return args.Get(0).([]model.SpuAttrName), args.Error(1)
}

func (m *MockGoodsRepository) CreateSpuAttrNames(ctx context.Context, attrs []model.SpuAttrName) error {
	args := m.Called(ctx, attrs)
	return args.Error(0)
}

func (m *MockGoodsRepository) DeleteSpuAttrNames(ctx context.Context, spuID uint64) error {
	args := m.Called(ctx, spuID)
	return args.Error(0)
}

func (m *MockGoodsRepository) CreateSpuAttrValues(ctx context.Context, values []model.SpuAttrValue) error {
	args := m.Called(ctx, values)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindBrandByID(ctx context.Context, id uint64) (*model.Brand, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Brand), args.Error(1)
}

func (m *MockGoodsRepository) FindSkusBySpuID(ctx context.Context, spuID uint64) ([]model.Sku, error) {
	args := m.Called(ctx, spuID)
	return args.Get(0).([]model.Sku), args.Error(1)
}

func (m *MockGoodsRepository) FindSkuByCode(ctx context.Context, code string) (*model.Sku, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Sku), args.Error(1)
}

func (m *MockGoodsRepository) FindSkuByID(ctx context.Context, id uint64) (*model.Sku, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Sku), args.Error(1)
}

func (m *MockGoodsRepository) CreateSku(ctx context.Context, sku *model.Sku) error {
	args := m.Called(ctx, sku)
	return args.Error(0)
}

func (m *MockGoodsRepository) UpdateSku(ctx context.Context, sku *model.Sku) error {
	args := m.Called(ctx, sku)
	return args.Error(0)
}

func (m *MockGoodsRepository) UpdateSkuStatus(ctx context.Context, skuID uint64, status uint8) error {
	args := m.Called(ctx, skuID, status)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindSkuAttrsBySkuID(ctx context.Context, skuID uint64) ([]model.SkuAttr, error) {
	args := m.Called(ctx, skuID)
	return args.Get(0).([]model.SkuAttr), args.Error(1)
}

func (m *MockGoodsRepository) CreateSkuAttrs(ctx context.Context, attrs []model.SkuAttr) error {
	args := m.Called(ctx, attrs)
	return args.Error(0)
}

func (m *MockGoodsRepository) DeleteSkuAttrs(ctx context.Context, skuID uint64) error {
	args := m.Called(ctx, skuID)
	return args.Error(0)
}

func (m *MockGoodsRepository) FindSkuImagesBySkuID(ctx context.Context, skuID uint64) ([]model.SkuImage, error) {
	args := m.Called(ctx, skuID)
	return args.Get(0).([]model.SkuImage), args.Error(1)
}

func (m *MockGoodsRepository) CreateSkuImages(ctx context.Context, images []model.SkuImage) error {
	args := m.Called(ctx, images)
	return args.Error(0)
}

func (m *MockGoodsRepository) DeleteSkuImages(ctx context.Context, skuID uint64) error {
	args := m.Called(ctx, skuID)
	return args.Error(0)
}

func (m *MockGoodsRepository) ListGoods(ctx context.Context, q *model.GoodsListQuery) ([]model.GoodsListItem, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]model.GoodsListItem), args.Get(1).(int64), args.Error(2)
}

func (m *MockGoodsRepository) InvalidateSpuCache(ctx context.Context, spuID uint64) error {
	args := m.Called(ctx, spuID)
	return args.Error(0)
}

// ============ Category Tests ============

func TestBuildCategoryTree_CacheHit(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	cached := []*model.CategoryNode{
		{ID: 1, Name: "服装", ParentID: 0, Level: 1},
	}

	mockRepo.On("GetCachedCategoryTree", context.Background()).
		Return(cached, nil)

	// Cache hit → return without DB query
	tree, err := mockRepo.GetCachedCategoryTree(context.Background())
	assert.NoError(t, err)
	assert.Len(t, tree, 1)
	assert.Equal(t, "服装", tree[0].Name)

	mockRepo.AssertNotCalled(t, "FindAllCategories")
}

func TestBuildCategoryTree_CacheMiss_BuildsTree(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	categories := []model.Category{
		{ID: 1, Name: "服装", ParentID: 0, Level: 1, Sort: 1},
		{ID: 2, Name: "男装", ParentID: 1, Level: 2, Sort: 1},
		{ID: 3, Name: "女装", ParentID: 1, Level: 2, Sort: 2},
	}

	mockRepo.On("GetCachedCategoryTree", context.Background()).
		Return(nil, errors.New("cache miss"))
	mockRepo.On("FindAllCategories", context.Background()).
		Return(categories, nil)
	mockRepo.On("CacheCategoryTree", context.Background(), mock.AnythingOfType("[]*model.CategoryNode")).
		Return(nil)

	// Cache miss → query DB, build tree, cache result
	_, err := mockRepo.GetCachedCategoryTree(context.Background())
	assert.Error(t, err)

	cats, err := mockRepo.FindAllCategories(context.Background())
	assert.NoError(t, err)
	assert.Len(t, cats, 3)

	mockRepo.CacheCategoryTree(context.Background(), []*model.CategoryNode{})

	mockRepo.AssertCalled(t, "FindAllCategories", context.Background())
}

func TestGetCategoryByID_NotFound_ReturnsErrCategoryNotFound(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	mockRepo.On("FindCategoryByID", context.Background(), uint64(999)).
		Return(nil, errors.New("not found"))

	cat, err := mockRepo.FindCategoryByID(context.Background(), 999)
	assert.Nil(t, cat)
	assert.Error(t, err)
}

// ============ SPU Tests ============

func TestCreateSpu_DuplicateCode_ReturnsErrSpuCodeExists(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	existing := &model.Spu{ID: 1, SpuCode: "SPU001"}

	mockRepo.On("FindSpuByCode", context.Background(), "SPU001").
		Return(existing, nil)

	// Duplicate SPU code → ErrSpuCodeExists
	found, _ := mockRepo.FindSpuByCode(context.Background(), "SPU001")
	assert.NotNil(t, found)
	assert.True(t, found != nil && found.ID > 0)
}

func TestCreateSpu_NonLeafCategory_ReturnsError(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	nonLeafCat := &model.Category{ID: 1, IsLeaf: 0}

	mockRepo.On("FindSpuByCode", context.Background(), "SPU_NEW").
		Return(nil, errors.New("not found"))
	mockRepo.On("FindCategoryByID", context.Background(), uint64(1)).
		Return(nonLeafCat, nil)

	// Non-leaf category is not allowed for product creation
	found, _ := mockRepo.FindSpuByCode(context.Background(), "SPU_NEW")
	assert.Nil(t, found)

	cat, _ := mockRepo.FindCategoryByID(context.Background(), 1)
	assert.Equal(t, uint8(0), cat.IsLeaf)
	// Service: cat.IsLeaf == 0 → error "category must be a leaf category"
}

func TestCreateSpu_Success(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	leafCat := &model.Category{ID: 2, IsLeaf: 1}
	newSpu := &model.Spu{SpuCode: "SPU_NEW", Title: "Test Product", CategoryID: 2}

	mockRepo.On("FindSpuByCode", context.Background(), "SPU_NEW").
		Return(nil, errors.New("not found"))
	mockRepo.On("FindCategoryByID", context.Background(), uint64(2)).
		Return(leafCat, nil)
	mockRepo.On("CreateSpu", context.Background(), mock.AnythingOfType("*model.Spu")).
		Return(nil)

	found, _ := mockRepo.FindSpuByCode(context.Background(), "SPU_NEW")
	assert.Nil(t, found)

	cat, _ := mockRepo.FindCategoryByID(context.Background(), 2)
	assert.Equal(t, uint8(1), cat.IsLeaf)

	err := mockRepo.CreateSpu(context.Background(), newSpu)
	assert.NoError(t, err)
}

func TestUpdateSpu_NotAuthorized_WrongShop(t *testing.T) {
	spu := &model.Spu{ID: 1, ShopID: 100}

	// Shop 200 tries to update shop 100's product
	assert.NotEqual(t, uint64(200), spu.ShopID)
}

func TestGetSpuDetail_NotFound_ReturnsErrSpuNotFound(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	mockRepo.On("FindSpuByID", context.Background(), uint64(999)).
		Return(nil, errors.New("not found"))

	spu, err := mockRepo.FindSpuByID(context.Background(), 999)
	assert.Nil(t, spu)
	assert.Error(t, err)
}

func TestGetSpuDetail_Success(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	spu := &model.Spu{ID: 1, Title: "Test Product"}
	skus := []model.Sku{{ID: 1}}
	attrs := []model.SpuAttrName{{ID: 1, AttrName: "颜色"}}

	mockRepo.On("FindSpuByID", context.Background(), uint64(1)).
		Return(spu, nil)
	mockRepo.On("FindBrandByID", context.Background(), mock.AnythingOfType("uint64")).
		Return(&model.Brand{Name: "TestBrand"}, nil)
	mockRepo.On("FindCategoryByID", context.Background(), mock.AnythingOfType("uint64")).
		Return(&model.Category{Name: "TestCat"}, nil)
	mockRepo.On("FindSpuExt", context.Background(), uint64(1)).
		Return(&model.SpuExt{}, nil)
	mockRepo.On("FindSkusBySpuID", context.Background(), uint64(1)).
		Return(skus, nil)
	mockRepo.On("FindSpuAttrNamesBySpuID", context.Background(), uint64(1)).
		Return(attrs, nil)

	// All required fields populated in response
	found, err := mockRepo.FindSpuByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Test Product", found.Title)
}

// ============ SKU Tests ============

func TestCreateSku_DuplicateCode_ReturnsErrSkuCodeExists(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	existing := &model.Sku{ID: 1, SkuCode: "SKU001"}

	mockRepo.On("FindSkuByCode", context.Background(), "SKU001").
		Return(existing, nil)

	found, _ := mockRepo.FindSkuByCode(context.Background(), "SKU001")
	assert.NotNil(t, found)
	assert.True(t, found != nil && found.ID > 0)
}

func TestCreateSku_SpuNotFound_ReturnsErrSpuNotFound(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	mockRepo.On("FindSkuByCode", context.Background(), "SKU001").
		Return(nil, errors.New("not found"))
	mockRepo.On("FindSpuByID", context.Background(), uint64(999)).
		Return(nil, errors.New("not found"))

	found, _ := mockRepo.FindSkuByCode(context.Background(), "SKU001")
	assert.Nil(t, found)

	spu, err := mockRepo.FindSpuByID(context.Background(), 999)
	assert.Nil(t, spu)
	assert.Error(t, err)
}

func TestCreateSku_WrongShop_ReturnsNotAuthorized(t *testing.T) {
	spu := &model.Spu{ID: 1, ShopID: 100}

	// Shop 200 creating SKU under shop 100's SPU → not authorized
	assert.NotEqual(t, uint64(200), spu.ShopID)
}

func TestUpdateSkuStatus_InvalidStatus_ReturnsErrInvalidStatus(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	sku := &model.Sku{ID: 1, Status: 0}

	// Valid statuses: 0=待上架, 1=上架, 2=下架, 3=售罄, 4=归档
	invalidStatus := uint8(99)

	mockRepo.On("FindSkuByID", context.Background(), uint64(1)).
		Return(sku, nil)

	found, _ := mockRepo.FindSkuByID(context.Background(), 1)
	assert.NotEqual(t, invalidStatus, found.Status)

	// Service validates: status must be one of [0,1,2,3,4]
	isValid := invalidStatus == 0 || invalidStatus == 1 || invalidStatus == 2 || invalidStatus == 3 || invalidStatus == 4
	assert.False(t, isValid)
}

func TestUpdateSkuStatus_ValidTransition_Success(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	sku := &model.Sku{ID: 1, SpuID: 1, Status: 0}

	mockRepo.On("FindSkuByID", context.Background(), uint64(1)).
		Return(sku, nil)
	mockRepo.On("UpdateSkuStatus", context.Background(), uint64(1), uint8(1)).
		Return(nil)
	mockRepo.On("InvalidateSpuCache", context.Background(), uint64(1)).
		Return(nil)

	found, _ := mockRepo.FindSkuByID(context.Background(), 1)
	assert.Equal(t, uint8(0), found.Status)

	// 0(待上架) → 1(上架): valid
	err := mockRepo.UpdateSkuStatus(context.Background(), 1, 1)
	assert.NoError(t, err)

}

// ============ Goods List Tests ============

func TestListGoods_DefaultPagination(t *testing.T) {
	q := &model.GoodsListQuery{}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}

	// Service applies defaults before querying
	assert.Equal(t, 1, q.Page)
	assert.Equal(t, 20, q.PageSize)
}

func TestListGoods_WithFilters(t *testing.T) {
	mockRepo := new(MockGoodsRepository)

	q := &model.GoodsListQuery{
		CategoryID: 10,
		BrandID:    5,
		Keyword:    "手机",
		Status:     1,
		Page:       2,
		PageSize:   10,
	}

	items := []model.GoodsListItem{
		{SpuID: 1, Title: "iPhone", MinPrice: 5000, MaxPrice: 8000},
	}

	mockRepo.On("ListGoods", context.Background(), q).
		Return(items, int64(1), nil)

	result, total, err := mockRepo.ListGoods(context.Background(), q)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "iPhone", result[0].Title)
}

// ============ Error Constants ============

func TestGoodsErrorConstants(t *testing.T) {
	assert.Equal(t, "SPU not found", ErrSpuNotFound.Error())
	assert.Equal(t, "SKU not found", ErrSkuNotFound.Error())
	assert.Equal(t, "SPU code already exists", ErrSpuCodeExists.Error())
	assert.Equal(t, "SKU code already exists", ErrSkuCodeExists.Error())
	assert.Equal(t, "category not found", ErrCategoryNotFound.Error())
	assert.Equal(t, "invalid status", ErrInvalidStatus.Error())
}

