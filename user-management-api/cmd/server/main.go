package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sevensolutions-backend/config"
	"sevensolutions-backend/internal/adapters/http/handler"
	"sevensolutions-backend/internal/adapters/http/router"
	mongoadapter "sevensolutions-backend/internal/adapters/mongo"
	"sevensolutions-backend/internal/application"
	"sevensolutions-backend/pkg/jwt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ctx ตัวเดียวคุมทั้ง DB, background counter, และ shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database(cfg.MongoDB).Collection("users")
	if err := mongoadapter.EnsureIndexes(ctx, collection); err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	repo := mongoadapter.NewMongoUserRepository(collection)
	jwtService := jwt.NewService(cfg.JWTSecret)
	authService := application.NewAuthService(repo, jwtService)
	userService := application.NewUserService(repo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	engine := router.New(authHandler, userHandler, jwtService)

	go logUserCountEvery10Seconds(ctx, userService)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	go func() {
		log.Printf("server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	waitForShutdown(cancel, server)
}

func logUserCountEvery10Seconds(ctx context.Context, userService *application.UserService) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count, err := userService.Count(ctx)
			if err != nil {
				log.Printf("failed to count users: %v", err)
				continue
			}
			log.Printf("total users: %d", count)
		case <-ctx.Done():
			return
		}
	}
}

// รอ SIGINT/SIGTERM แล้วให้เวลา request ที่ค้างอยู่ 5 วิ ก่อนปิด
func waitForShutdown(cancel context.CancelFunc, server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
