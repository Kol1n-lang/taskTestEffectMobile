package main

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "taskTestEffectMobile/docs"
	"taskTestEffectMobile/internal/core/configs"
	"taskTestEffectMobile/internal/core/database"
	"taskTestEffectMobile/internal/handler"
	"taskTestEffectMobile/internal/repository"
	"taskTestEffectMobile/internal/service"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func initRouters(app *http.ServeMux, handler *handler.SubscriptionHandler) {
	handler.CreateSubscriptionsRoutes(app)
	log.Println("Router initialized")
}

// @title Subscription API
// @version 1.0
// @description API для управления подписками
// @host localhost:8080
// @BasePath /api/v1
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("can't sync zap logger: %v", err)
		}
	}()

	cfg := configs.Init()
	err = database.RunMigrations(cfg.DB.DBUrl())
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.CreateDBConnection()
	if err != nil {
		log.Fatal(err)
	}
	app := http.NewServeMux()

	subscriptionRepo := repository.NewSubscriptionRepository(db, logger)
	subscriptionService := service.NewSubscriptionService(*subscriptionRepo, logger)
	subscriptionHandler := handler.NewSubscriptionHandler(*subscriptionService, logger)

	initRouters(app, subscriptionHandler)
	app.Handle("/swagger/", httpSwagger.WrapHandler)
	handlerWithCORS := enableCORS(app)

	err = http.ListenAndServe(":8080", handlerWithCORS)
	if err != nil {
		log.Fatal(err)
	}
}
