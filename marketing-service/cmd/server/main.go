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

	"github.com/ecommerce/marketing-service/internal/config"
	"github.com/ecommerce/marketing-service/internal/handler"
	"github.com/ecommerce/marketing-service/internal/model"
	"github.com/ecommerce/marketing-service/internal/repository"
	"github.com/ecommerce/marketing-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password, DB: cfg.Redis.DB,
	})

	repo := repository.NewMarketingRepository(db, rdb)
	svc := service.NewMarketingService(repo)
	marketingHandler := handler.NewMarketingHandler(svc)

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(security.Recovery())
	router.Use(security.RequestLog(nil))
	// security middleware registered above
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	marketingHandler.RegisterRoutes(router)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Server.Port), Handler: router}
	go func() {
		log.Printf("marketing-service starting on :%s\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Warn)})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	db.AutoMigrate(&model.Coupon{}, &model.UserCoupon{}, &model.Promotion{}, &model.PromotionSku{})
	return db, nil
}
