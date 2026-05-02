package model

// 请求
type RegisterShopRequest struct {
	Name         string  `json:"name" binding:"required,max=64"`
	Type        uint8   `json:"type" binding:"required,oneof=1 2 3 4"`
	Province    string  `json:"province" binding:"required"`
	City        string  `json:"city" binding:"required"`
	Description string  `json:"description"`
}

type UpdateShopRequest struct {
	Name         string `json:"name"`
	BannerURL   string `json:"banner_url"`
	Description string `json:"description"`
}

type CreateFreightTemplateRequest struct {
	Name             string `json:"name" binding:"required"`
	Type            uint8  `json:"type" binding:"required,oneof=1 2 3 4"`
	IsFreeThreshold uint8  `json:"is_free_threshold"`
	FreeAmount      float64 `json:"free_amount"`
	FreeNum         int     `json:"free_num"`
	Rules           []FreightRuleDTO `json:"rules"`
}

type FreightRuleDTO struct {
	ProvinceCodes string  `json:"province_codes"`
	FirstAmount  float64 `json:"first_amount"`
	FirstNum     int     `json:"first_num"`
	AddAmount    float64 `json:"add_amount"`
	AddNum       int     `json:"add_num"`
}

type UpdateDecorationRequest struct {
	Layout string `json:"layout" binding:"required"` // JSON
}

// 响应
type ShopResponse struct {
	Shop *Shop `json:"shop"`
}

type FreightTemplateResponse struct {
	Template *FreightTemplate `json:"template"`
	Rules    []FreightRule    `json:"rules"`
}
