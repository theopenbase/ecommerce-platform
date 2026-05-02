package model

// ============ 请求 DTO ============

type RegisterRequest struct {
	Mobile   string `json:"mobile" binding:"required,len=11"`
	Nickname string `json:"nickname" binding:"required,min=2,max=32"`
	Code     string `json:"code" binding:"required,len=6"`
}

type LoginRequest struct {
	Mobile string `json:"mobile" binding:"required,len=11"`
	Code   string `json:"code" binding:"required,len=6"`
}

type SendCodeRequest struct {
	Mobile string `json:"mobile" binding:"required,len=11"`
}

type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"omitempty,min=2,max=32"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
	Gender    uint8  `json:"gender" binding:"omitempty,oneof=0 1 2"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

type SetPayPasswordRequest struct {
	PayPassword string `json:"pay_password" binding:"required,min=6,max=32"`
	Code        string `json:"code" binding:"required,len=6"`
}

type AddressRequest struct {
	Receiver      string `json:"receiver" binding:"required,min=1,max=32"`
	Mobile       string `json:"mobile" binding:"required,len=11"`
	ProvinceCode string `json:"province_code" binding:"required"`
	CityCode     string `json:"city_code" binding:"required"`
	DistrictCode string `json:"district_code" binding:"required"`
	ProvinceName string `json:"province_name" binding:"required"`
	CityName     string `json:"city_name" binding:"required"`
	DistrictName string `json:"district_name" binding:"required"`
	Detail       string `json:"detail" binding:"required,min=1,max=256"`
	Tag          string `json:"tag" binding:"omitempty,max=16"`
	IsDefault    uint8  `json:"is_default" binding:"omitempty,oneof=0 1"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ============ 响应 DTO ============

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserResponse struct {
	User   *UserProfile  `json:"user"`
	Tokens *TokenResponse `json:"tokens,omitempty"`
}

type AddressResponse struct {
	List []ReceiverAddress `json:"list"`
}

type MemberResponse struct {
	Member   *MemberProfile `json:"member"`
	LevelInfo *MemberLevel  `json:"level_info,omitempty"`
}

type PointsResponse struct {
	Account *PointsAccount `json:"account"`
}
