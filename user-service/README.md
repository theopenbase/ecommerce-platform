# 用户模块（user-service）

Go + Gin + GORM + Redis，遵循"约定优于配置"原则。

## 目录结构

```
user-service/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── config/                 # 配置
│   ├── handler/                # HTTP Handler
│   ├── service/                # 业务逻辑
│   ├── repository/             # 数据访问层
│   ├── model/                 # 数据模型 & DTO
│   ├── middleware/            # JWT 认证中间件
│   └── pkg/
│       ├── jwt/                # JWT Token 管理
│       └── smc/                # 短信验证码接口
├── migrations/                 # SQL 迁移脚本
└── go.mod
```

## 核心能力

### 1. 认证

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/users/send_code` | POST | 发送短信验证码（Redis 5min TTL）|
| `/api/v1/users/register` | POST | 用户注册（手机+验证码）|
| `/api/v1/users/login` | POST | 登录（手机+验证码）|
| `/POST /api/v1/users/refresh_token` | POST | 刷新 AccessToken |
| `/api/v1/users/logout` | POST | 登出（作废 RefreshToken）|

### 2. 个人资料

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/users/me` | GET | 获取个人资料 |
| `/api/v1/users/me` | PUT | 更新个人资料（昵称/头像/性别）|
| `/api/v1/users/password` | POST | 修改登录密码 |
| `/api/v1/users/pay_password` | POST | 设置支付密码（验证码确认）|
| `/api/v1/users/login_logs` | GET | 获取登录日志（最近20条）|

### 3. 收货地址

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/users/addresses` | GET | 获取地址列表 |
| `/api/v1/users/addresses` | POST | 新增地址 |
| `/api/v1/users/addresses/:id` | PUT | 更新地址 |
| `/api/v1/users/addresses/:id` | DELETE | 删除地址 |
| `/api/v1/users/addresses/:id/default` | PUT | 设置默认地址 |

**限制**：每人最多 20 个收货地址。

### 4. 会员中心

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/users/member/profile` | GET | 获取会员等级/成长值/下一级门槛 |

### 5. 积分账户

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/users/points/balance` | GET | 获取积分余额和累计信息 |

## JWT Token 设计

- AccessToken：15分钟，承载 user_id
- RefreshToken：7天，存储于 Redis，Logout 时从 Redis 删除
- Token 作废：Redis SETNX，Logout 时删除 Key

## 短信验证码

- 存储：Redis Key = `sms:code:{mobile}`，TTL = 5 分钟
- 生产集成：阿里云 SMS / 腾讯云 SMS（MockSMSProvider）

## 数据库

- MySQL InnoDB，字符集 utf8mb4
- GORM AutoMigrate 自动建表

## 启动

```bash
go mod tidy
go run cmd/server/main.go
```

环境变量：

| 变量 | 默认值 |
|------|--------|
| SERVER_PORT | 8080 |
| DB_HOST | localhost |
| DB_PORT | 3306 |
| DB_USER | root |
| DB_PASSWORD | |
| DB_NAME | ecommerce_user |
| REDIS_HOST | localhost |
| REDIS_PORT | 6379 |
| JWT_SECRET | your-256bit-secret |
