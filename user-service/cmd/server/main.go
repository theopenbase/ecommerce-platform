package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ecommerce/common/middleware/security"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ecommerce/user-service/internal/config"
	"github.com/ecommerce/user-service/internal/handler"
	"github.com/ecommerce/user-service/internal/middleware"
	"github.com/ecommerce/user-service/internal/model"
	"github.com/ecommerce/user-service/internal/pkg/jwt"
	"github.com/ecommerce/user-service/internal/pkg/smc"
	"github.com/ecommerce/user-service/internal/repository"
	"github.com/ecommerce/user-service/internal/service"
)

func main() {
	cfg := config.Load()

	// 初始化数据库
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	// 初始化 Redis
	rdb := initRedis(cfg)

	// 初始化组件
	jwtMgr := jwt.NewManager(cfg.JWT.Secret)
	smsProv := &smc.MockSMSProvider{}
	repo := repository.NewUserRepository(db, rdb)
	svc := service.NewUserService(repo, jwtMgr, smsProv, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	userHandler := handler.NewUserHandler(svc)
	authMW := middleware.NewAuthMiddleware(jwtMgr)

	// 初始化 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(security.Recovery())
	router.Use(security.RequestLog(nil))
	// security middleware registered above

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 注册路由
	userHandler.RegisterRoutes(router, authMW)

	// 启动服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		log.Printf("server starting on :%s\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}
	log.Println("server exited")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := db.AutoMigrate(
		&model.User{},
		&model.LoginLog{},
		&model.Member{},
		&model.MemberLevel{},
		&model.PointsAccount{},
		&model.PointsLog{},
		&model.ReceiverAddress{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
