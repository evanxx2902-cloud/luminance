package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/lib/pq"

	"github.com/luminance/backend/internal/biz"
	"github.com/luminance/backend/internal/data"
	"github.com/luminance/backend/internal/server"
	"github.com/luminance/backend/internal/service"
)

func main() {
	logger := log.NewStdLogger(os.Stdout)
	helper := log.NewHelper(logger)

	// 初始化数据库连接
	db, err := sql.Open("postgres", getDatabaseURL())
	if err != nil {
		helper.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		helper.Fatalf("failed to ping database: %v", err)
	}

	// 初始化 Redis 客户端
	redisClient := data.NewRedisClient(getRedisAddr())
	if err := redisClient.Ping(context.Background()); err != nil {
		helper.Warnf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// 初始化 Repository
	userRepo := data.NewUserRepo(db)

	// 初始化认证器
	authRegistry := biz.NewAuthRegistry()
	passwordAuth := biz.NewPasswordAuthenticator(userRepo)
	authRegistry.Register(passwordAuth)

	// 初始化 UseCase
	userUseCase := biz.NewUserUseCase(userRepo, redisClient, authRegistry)

	// 初始化 Service
	userService := service.NewUserService(userUseCase)

	// 创建服务器
	httpSrv := server.NewHTTPServer(userService)
	grpcSrv := server.NewGRPCServer()

	app := kratos.New(
		kratos.Name("luminance-api"),
		kratos.Logger(logger),
		kratos.Server(httpSrv, grpcSrv),
	)

	if err := app.Run(); err != nil {
		helper.Fatalf("failed to run luminance-api: %v", err)
	}
}

// getDatabaseURL 获取数据库连接字符串
func getDatabaseURL() string {
	if url := os.Getenv("DATABASE_BUSINESS_URL"); url != "" {
		return url
	}
	return "postgres://luminance:luminance@localhost:5432/luminance?sslmode=disable"
}

// getRedisAddr 获取 Redis 地址
func getRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}
