#!/bin/bash
# =============================================================================
# Phase 4 — R-3 退款支付密码阈值验证
# 执行环境：CI docker-compose 或 K8s deployment
# 前置条件：Charlie 实现 VerifyPayPassword 中间件
# =============================================================================

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN=""
USER_ID=""

# -----------------------------------------------------------------------------
# 工具函数
# -----------------------------------------------------------------------------

get_token() {
  local mobile="$1"
  local code="${2:-123456}"
  curl -s -X POST "$BASE_URL/api/v1/users/send_code" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\"}" > /dev/null
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/users/login" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\",\"code\":\"$code\"}")
  USER_ID=$(echo "$resp" | grep -o '"user_id":[0-9]*' | cut -d':' -f2)
  echo "[AUTH] user_id=$USER_ID"
}

# 创建测试订单（需先有可退款的订单）
create_test_order() {
  local user_id="$1"
  local sku_id="${2:-1}"
  local qty="${3:-1}"

  # 添加购物车
  curl -s -X POST "$BASE_URL/api/v1/cart/items" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $user_id" \
    -d "{\"sku_id\":$sku_id,\"quantity\":$qty}" > /dev/null

  CART_ITEM_ID=$(curl -s -X GET "$BASE_URL/api/v1/cart" \
    -H "X-User-ID: $user_id" \
    | grep -o '"cart_id":[0-9]*' | head -1 | cut -d':' -f2)

  # 创建订单
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $user_id" \
    -d "{\"address_id\":1,\"items\":[$CART_ITEM_ID]}")

  echo "$resp"
}

# -----------------------------------------------------------------------------
# PHASE 4 — R-3 退款支付密码阈值测试
# Bob产品决策：≤200元仅登录验证，>200元强制验支付密码
# -----------------------------------------------------------------------------

echo ""
echo "=============================================="
echo "PHASE 4: R-3 退款支付密码阈值测试"
echo "=============================================="


# -----------------------------------------------------------------------------
# T-7a: 小额退款（≤200元）无支付密码 → 通过
# -----------------------------------------------------------------------------
echo ""
echo "[T-7a] 小额退款（≤200元）无支付密码 → 预期：通过"

get_token "13800138101" "123456"
ORDER_RESP=$(create_test_order "$USER_ID" 1 1)
ORDER_NO=$(echo "$ORDER_RESP" | grep -o '"order_no":"[^"]*' | cut -d'"' -f4)
SUB_ORDER_NO=$(echo "$ORDER_RESP" | grep -o '"sub_order_no":"[^"]*' | cut -d'"' -f4 | head -1)

echo "[T-7a] user_id=$USER_ID order_no=$ORDER_NO sub_order_no=$SUB_ORDER_NO"

# 模拟付款（实际需要支付成功）
# 然后申请小额退款（≤200元）
REFUND_RESP=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":150.00}")

REFUND_CODE=$(echo "$REFUND_RESP" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-7a] 退款响应 code=$REFUND_CODE（预期：0=成功）"


# -----------------------------------------------------------------------------
# T-7b: 0元退款（仅退优惠券场景）无支付密码 → 通过
# -----------------------------------------------------------------------------
echo ""
echo "[T-7b] 0元退款无支付密码 → 预期：通过"

REFUND_RESP2=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"仅退优惠券\",\"amount\":0.00}")

REFUND_CODE2=$(echo "$REFUND_RESP2" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-7b] 退款响应 code=$REFUND_CODE2（预期：0=成功）"


# -----------------------------------------------------------------------------
# T-8a: 大额退款（>200元）无支付密码 → 拒绝
# -----------------------------------------------------------------------------
echo ""
echo "[T-8a] 大额退款（>200元）无支付密码 → 预期：拒绝（ErrPayPasswordNotSet）"

REFUND_RESP3=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":500.00}")

REFUND_CODE3=$(echo "$REFUND_RESP3" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-8a] 退款响应 code=$REFUND_CODE3（预期：非0=拒绝）"


# -----------------------------------------------------------------------------
# T-8b: 大额退款（>200元）支付密码错误 → 拒绝
# -----------------------------------------------------------------------------
echo ""
echo "[T-8b] 大额退款（>200元）支付密码错误 → 预期：拒绝（ErrIncorrectPayPassword）"

REFUND_RESP4=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":500.00,\"pay_password\":\"000000\"}")

REFUND_CODE4=$(echo "$REFUND_RESP4" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-8b] 退款响应 code=$REFUND_CODE4（预期：非0=拒绝）"


# -----------------------------------------------------------------------------
# T-9: 大额退款（>200元）支付密码正确 → 受理成功
# -----------------------------------------------------------------------------
echo ""
echo "[T-9] 大额退款（>200元）支付密码正确 → 预期：受理成功"

# 先确保用户设置了支付密码（通过接口设置测试密码）
curl -s -X POST "$BASE_URL/api/v1/users/pay_password" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"pay_password\":\"123456\",\"code\":\"123456\"}" > /dev/null

REFUND_RESP5=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":500.00,\"pay_password\":\"123456\"}")

REFUND_CODE5=$(echo "$REFUND_RESP5" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-9] 退款响应 code=$REFUND_CODE5（预期：0=成功）"


# -----------------------------------------------------------------------------
# T-9a: seller 审批退款（任意金额）无需支付密码
# -----------------------------------------------------------------------------
echo ""
echo "[T-9a] seller 审批退款（任意金额）→ 预期：无密码要求，直接处理"

SELLER_ID="13800138102"  # seller 账号
# seller 审批接口（不同路径，seller角色审批）
SELLER_RESP=$(curl -s -X POST "$BASE_URL/api/v1/seller/refund/approve" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $SELLER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"action\":\"approve\"}")

SELLER_CODE=$(echo "$SELLER_RESP" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-9a] seller审批响应 code=$SELLER_CODE（预期：0=成功，无密码验证）"


# -----------------------------------------------------------------------------
# T-9b: 边界值 200元退款 → 无需支付密码
# -----------------------------------------------------------------------------
echo ""
echo "[T-9b] 边界值：200元退款 → 预期：无需支付密码，通过"

REFUND_RESP6=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":200.00}")

REFUND_CODE6=$(echo "$REFUND_RESP6" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-9b] 退款响应 code=$REFUND_CODE6（预期：0=成功，200≤200免密）"


# -----------------------------------------------------------------------------
# T-9c: 边界值 201元退款 → 需要支付密码
# -----------------------------------------------------------------------------
echo ""
echo "[T-9c] 边界值：201元退款 → 预期：需要支付密码"

REFUND_RESP7=$(curl -s -X POST "$BASE_URL/api/v1/orders/$ORDER_NO/refund" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d "{\"sub_order_no\":\"$SUB_ORDER_NO\",\"type\":1,\"reason\":\"商品损坏\",\"amount\":201.00}")

REFUND_CODE7=$(echo "$REFUND_RESP7" | grep -o '"code":[0-9]*' | cut -d':' -f2)
echo "[T-9c] 退款响应 code=$REFUND_CODE7（预期：非0=拒绝，201>200需验密）"


echo ""
echo "=============================================="
echo "Phase 4 完成 — 结果汇总"
echo "=============================================="
echo "T-7a: ≤200元无密码 → code=$REFUND_CODE（预期0）"
echo "T-7b: 0元无密码    → code=$REFUND_CODE2（预期0）"
echo "T-8a: >200元无密码 → code=$REFUND_CODE3（预期非0）"
echo "T-8b: >200元密码错 → code=$REFUND_CODE4（预期非0）"
echo "T-9:  >200元密码对 → code=$REFUND_CODE5（预期0）"
echo "T-9a: seller审批   → code=$SELLER_CODE（预期0）"
echo "T-9b: 200元边界    → code=$REFUND_CODE6（预期0）"
echo "T-9c: 201元边界    → code=$REFUND_CODE7（预期非0）"
