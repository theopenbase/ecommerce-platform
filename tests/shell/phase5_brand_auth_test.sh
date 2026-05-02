#!/bin/bash
# =============================================================================
# Phase 5 — R-6 品牌授权类目资质验证
# 执行环境：CI docker-compose 或 K8s deployment
# 前置条件：Charlie 实现类目资质读取 + brand_authorization 凭证校验
# =============================================================================

BASE_URL="${BASE_URL:-http://localhost:8080}"
SHOP_TOKEN=""
SHOP_ID=""

# -----------------------------------------------------------------------------
# 工具函数
# -----------------------------------------------------------------------------

shop_login() {
  local mobile="$1"
  local code="${2:-123456}"
  curl -s -X POST "$BASE_URL/api/v1/users/send_code" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\"}" > /dev/null
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/users/login" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\",\"code\":\"$code\"}")
  SHOP_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
  SHOP_ID=$(echo "$resp" | grep -o '"user_id":[0-9]*' | cut -d':' -f2)
  echo "[SHOP_AUTH] shop_id=$SHOP_ID"
}

# -----------------------------------------------------------------------------
# PHASE 5 — R-6 品牌授权类目资质测试
# Bob产品决策：读取末级类目资质要求，含brand_authorization则必须上传凭证
# -----------------------------------------------------------------------------

echo ""
echo "=============================================="
echo "PHASE 5: R-6 品牌授权类目资质测试"
echo "=============================================="


# -----------------------------------------------------------------------------
# T-14a: 有资质要求类目 + 上传授权凭证 → 通过
# -----------------------------------------------------------------------------
echo ""
echo "[T-14a] 有资质要求类目 + 有品牌授权凭证 → 预期：通过"

shop_login "13800138201" "123456"

# 获取有品牌授权要求的末级类目
CATEGORY_WITH_AUTH=$(curl -s -X GET "$BASE_URL/api/v1/categories/1" \
  -H "X-User-ID: $SHOP_ID" \
  | grep -o '"has_brand_auth_required":[^,]*' | cut -d':' -f2)

echo "[T-14a] 类目品牌授权要求=$CATEGORY_WITH_AUTH"

# 创建 SPU（带品牌授权凭证）
SPU_RESP=$(curl -s -X POST "$BASE_URL/api/v1/seller/products" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SHOP_ID" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d "{
    \"spu_code\": \"SPU-AUTH-001\",
    \"title\": \"品牌授权商品\",
    \"short_desc\": \"Test product with brand auth\",
    \"brand_id\": 1,
    \"category_id\": 1,
    \"unit\": \"件\",
    \"brand_authorization\": \"base64_encoded_cert_here\"
  }")

SPU_CODE=$(echo "$SPU_RESP" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-14a] 创建SPU响应 code=$SPU_CODE（预期：0=成功）"


# -----------------------------------------------------------------------------
# T-14b: 有资质要求类目 + 无授权凭证 → 阻断
# -----------------------------------------------------------------------------
echo ""
echo "[T-14b] 有资质要求类目 + 无品牌授权凭证 → 预期：阻断（ErrBrandAuthRequired）"

SPU_RESP2=$(curl -s -X POST "$BASE_URL/api/v1/seller/products" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SHOP_ID" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d "{
    \"spu_code\": \"SPU-NO-AUTH-001\",
    \"title\": \"无品牌授权商品\",
    \"short_desc\": \"Test product without brand auth\",
    \"brand_id\": 1,
    \"category_id\": 1,
    \"unit\": \"件\",
    \"brand_authorization\": \"\"
  }")

SPU_CODE2=$(echo "$SPU_RESP2" | grep -o '"code":[0-9]*' | cut -d':' -f2)
SPU_MSG=$(echo "$SPU_RESP2" | grep -o '"message":"[^"]*' | cut -d'"' -f4)
echo "[T-14b] 创建SPU响应 code=$SPU_CODE2 msg=$SPU_MSG（预期：非0，ErrBrandAuthRequired）"


# -----------------------------------------------------------------------------
# T-14c: 无资质要求类目 + 无授权凭证 → 通过
# -----------------------------------------------------------------------------
echo ""
echo "[T-14c] 无资质要求类目 + 无品牌授权凭证 → 预期：通过"

# 获取无品牌授权要求的末级类目
CATEGORY_NO_AUTH=$(curl -s -X GET "$BASE_URL/api/v1/categories/10" \
  -H "X-User-ID: $SHOP_ID" \
  | grep -o '"has_brand_auth_required":[^,]*' | cut -d':' -f2)

echo "[T-14c] 类目品牌授权要求=$CATEGORY_NO_AUTH"

SPU_RESP3=$(curl -s -X POST "$BASE_URL/api/v1/seller/products" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SHOP_ID" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d "{
    \"spu_code\": \"SPU-NO-REQ-001\",
    \"title\": \"无需品牌授权商品\",
    \"short_desc\": \"Test product without brand auth requirement\",
    \"brand_id\": 1,
    \"category_id\": 10,
    \"unit\": \"件\"
  }")

SPU_CODE3=$(echo "$SPU_RESP3" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-14c] 创建SPU响应 code=$SPU_CODE3（预期：0=成功）"


# -----------------------------------------------------------------------------
# T-14d: 有资质要求类目 + 多文件部分缺失 → 阻断
# -----------------------------------------------------------------------------
echo ""
echo "[T-14d] 有资质要求类目（多文件） + 部分文件缺失 → 预期：阻断"

# 假设类目要求 brand_authorization + business_license，但只上传了 brand_authorization
SPU_RESP4=$(curl -s -X POST "$BASE_URL/api/v1/seller/products" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SHOP_ID" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d "{
    \"spu_code\": \"SPU-PARTIAL-AUTH-001\",
    \"title\": \"部分资质商品\",
    \"short_desc\": \"Test product with partial docs\",
    \"brand_id\": 1,
    \"category_id\": 1,
    \"unit\": \"件\",
    \"brand_authorization\": \"base64_brand_auth_cert\",
    \"business_license\": \"\"
  }")

SPU_CODE4=$(echo "$SPU_RESP4" | grep -o '"code":[0-9]*' | cut -d':' -f2)
SPU_MSG4=$(echo "$SPU_RESP4" | grep -o '"message":"[^"]*' | cut -d'"' -f4)
echo "[T-14d] 创建SPU响应 code=$SPU_CODE4 msg=$SPU_MSG4（预期：非0，部分文件缺失）"


# -----------------------------------------------------------------------------
# T-14e: 非末级类目有资质要求 → 不触发校验
# -----------------------------------------------------------------------------
echo ""
echo "[T-14e] 非末级类目有资质要求 → 预期：不触发校验，正常创建"

# 非末级类目（如父类目）不应触发品牌授权检查
SPU_RESP5=$(curl -s -X POST "$BASE_URL/api/v1/seller/products" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SHOP_ID" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d "{
    \"spu_code\": \"SPU-PARENT-CAT-001\",
    \"title\": \"父类目商品\",
    \"short_desc\": \"Test product on parent category\",
    \"brand_id\": 1,
    \"category_id\": 0,
    \"unit\": \"件\"
  }")

SPU_CODE5=$(echo "$SPU_RESP5" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-14e] 创建SPU响应 code=$SPU_CODE5（预期：0=成功，不触发品牌授权检查）"


echo ""
echo "=============================================="
echo "Phase 5 完成 — 结果汇总"
echo "=============================================="
echo "T-14a: 有要求+有凭证 → code=$SPU_CODE（预期0）"
echo "T-14b: 有要求+无凭证 → code=$SPU_CODE2（预期非0）"
echo "T-14c: 无要求+无凭证 → code=$SPU_CODE3（预期0）"
echo "T-14d: 有要求+部分文件 → code=$SPU_CODE4（预期非0）"
echo "T-14e: 非末级类目   → code=$SPU_CODE5（预期0）"
