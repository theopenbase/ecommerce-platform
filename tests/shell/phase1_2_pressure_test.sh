#!/bin/bash
# =============================================================================
# Phase 1 & Phase 2 — 集成压测脚本
# 执行环境：CI docker-compose 或 K8s deployment
# 前置条件：user-service / goods-service / order-service 已部署
# =============================================================================

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN=""   # 登录后获取
USER_ID="" # X-User-ID header

# -----------------------------------------------------------------------------
# 工具函数
# -----------------------------------------------------------------------------

get_token() {
  local mobile="$1"  # e.g. "13800138001"
  local code="${2:-123456}"

  # 发送验证码
  curl -s -X POST "$BASE_URL/api/v1/users/send_code" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\"}" > /dev/null

  # 登录获取 token
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/users/login" \
    -H "Content-Type: application/json" \
    -d "{\"mobile\":\"$mobile\",\"code\":\"$code\"}")

  TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
  USER_ID=$(echo "$resp" | grep -o '"user_id":[0-9]*' | cut -d':' -f2)
  echo "[AUTH] token=$TOKEN user_id=$USER_ID"
}

# 添加购物车项（需先有商品 SKU）
add_cart() {
  local sku_id="$1"
  local qty="${2:-1}"

  curl -s -X POST "$BASE_URL/api/v1/cart/items" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $USER_ID" \
    -d "{\"sku_id\":$sku_id,\"quantity\":$qty}"
}

# -----------------------------------------------------------------------------
# PHASE 1 — R-4 幂等性测试
# 目标：10个并发请求，相同用户+相同购物车Items → 仅1单成功，其余返回 ErrIdempotentKeyUsed
# 预期：10个请求中，9个返回 code=3003（订单已存在），1个返回 code=0（成功）
# -----------------------------------------------------------------------------

echo ""
echo "=============================================="
echo "PHASE 1: R-4 幂等性并发测试"
echo "=============================================="

# 准备：登录 + 添加购物车
get_token "13800138001" "123456"
SKU_ID=1  # 替换为实际可用 SKU ID
add_cart $SKU_ID 2 > /dev/null

# 获取购物车第一个 item id
CART_ITEM_ID=$(curl -s -X GET "$BASE_URL/api/v1/cart" \
  -H "X-User-ID: $USER_ID" \
  | grep -o '"cart_id":[0-9]*' | head -1 | cut -d':' -f2)

echo "[PHASE1] user_id=$USER_ID, cart_item_id=$CART_ITEM_ID"
echo "[PHASE1] 发起10个并发创建订单请求..."

# 并发请求函数
create_order_once() {
  local idx="$1"
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $USER_ID" \
    -d "{\"address_id\":1,\"items\":[$CART_ITEM_ID]}")

  local code=$(echo "$resp" | grep -o '"code":[0-9]*' | cut -d':' -f2)
  local msg=$(echo "$resp" | grep -o '"message":"[^"]*' | cut -d'"' -f4)
  echo "[REQ-$idx] code=$code msg=$msg"
}

export -f create_order_once
export BASE_URL USER_ID CART_ITEM_ID

# 10并发
for i in $(seq 1 10); do
  create_order_once $i &
done
wait

echo "[PHASE1] 预期：1个code=0，9个code=3003"
echo "[PHASE1] 验证：统计成功/重复请求数量"


# -----------------------------------------------------------------------------
# PHASE 2 — R-1 库存超卖测试
# 目标：SKU库存=3，并发10个请求每个扣减2件 → 仅1单成功
# 预期：10个请求中，成功数 ≤ floor(3/2)=1
# -----------------------------------------------------------------------------

echo ""
echo "=============================================="
echo "PHASE 2: R-1 库存超卖并发测试"
echo "=============================================="

# 准备新用户 + 同一 SKU
get_token "13800138002" "123456"
SKU_ID_STOCK=2  # 替换为实际低库存 SKU ID
QTY=2

# 清空该用户的购物车后添加
curl -s -X DELETE "$BASE_URL/api/v1/cart/items/$CART_ITEM_ID" \
  -H "X-User-ID: $USER_ID" > /dev/null

add_cart $SKU_ID_STOCK $QTY > /dev/null
CART_ITEM_ID2=$(curl -s -X GET "$BASE_URL/api/v1/cart" \
  -H "X-User-ID: $USER_ID" \
  | grep -o '"cart_id":[0-9]*' | head -1 | cut -d':' -f2)

echo "[PHASE2] user_id=$USER_ID, sku_id=$SKU_ID_STOCK, qty=$QTY"
echo "[PHASE2] 库存预设为3件，并发10个请求..."

# 并发下单扣库存
create_order_stock() {
  local idx="$1"
  local resp
  resp=$(curl -s -X POST "$BASE_URL/api/v1/orders" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: $USER_ID" \
    -d "{\"address_id\":1,\"items\":[$CART_ITEM_ID2]}")

  local code=$(echo "$resp" | grep -o '"code":[0-9]*' | cut -d':' -f2)
  local order_no=$(echo "$resp" | grep -o '"order_no":"[^"]*' | cut -d'"' -f4)
  echo "[REQ-$idx] code=$code order_no=$order_no"
}

export -f create_order_stock
export BASE_URL USER_ID CART_ITEM_ID2

for i in $(seq 1 10); do
  create_order_stock $i &
done
wait

echo "[PHASE2] 预期：成功数 ≤ floor(3/2)=1（超卖 = 成功数 > 1）"
echo "[PHASE2] 验证：对比 Redis 库存扣减记录与实际成功订单数"


# -----------------------------------------------------------------------------
# 结果汇总
# -----------------------------------------------------------------------------
echo ""
echo "=============================================="
echo "压测完成 — 请核对上述输出"
echo "=============================================="
echo "Phase1 幂等性：预期 1 成功 + 9 duplicate"
echo "Phase2 库存超卖：预期 ≤1 成功（若 >1 则存在超卖风险）"
