package goodsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GoodsClient 调用 goods-service 获取商品快照信息
// 生产环境应使用连接池管理 HTTP Client
type GoodsClient struct {
	baseURL string
	hc      *http.Client
}

// SkuSnapshot SKU 关键字段快照（用于下单时保存商品信息）
type SkuSnapshot struct {
	SkuID     uint64  `json:"id"`
	SpuID     uint64  `json:"spu_id"`
	ShopID    uint64  `json:"shop_id"`
	SkuCode   string  `json:"sku_code"`
	Title     string  `json:"title"`
	SkuAttrs  string  `json:"sku_attrs"` // JSON 格式属性
	PriceTag  float64 `json:"price_tag"`
	PriceSell float64 `json:"price_sell"`
}

func NewGoodsClient(goodsHost string) *GoodsClient {
	return &GoodsClient{
		baseURL: fmt.Sprintf("http://%s", goodsHost),
		hc:      &http.Client{Timeout: 3 << 20}, // 3s timeout
	}
}

// GetSkuSnapshot 获取 SKU 快照（用于下单前填充购物车快照字段）
func (c *GoodsClient) GetSkuSnapshot(ctx context.Context, skuID uint64) (*SkuSnapshot, error) {
	url := fmt.Sprintf("%s/api/v1/goods/skus/%d/snapshot", c.baseURL, skuID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("goods service returned %d", resp.StatusCode)
	}
	var result struct {
		Code int          `json:"code"`
		Data *SkuSnapshot `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 || result.Data == nil {
		return nil, fmt.Errorf("failed to get sku snapshot")
	}
	return result.Data, nil
}
