package main

import (
	"backend/internal/app/config"
	"backend/internal/app/dsn"
	"backend/internal/app/handler"
	"backend/internal/app/middleware"
	"backend/internal/app/repository"
	"backend/internal/app/service"
	"backend/internal/pkg"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "backend/docs" // Swagger docs
)

// @title Flight Request API
// @version 1.0
// @description API системы управления заявками на полеты с JWT аутентификацией
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	router := gin.Default()

	// CORS middleware - должен быть ПЕРВЫМ
	router.Use(func(c *gin.Context) {
		fmt.Printf("=== CORS MIDDLEWARE ===\n")
		fmt.Printf("Method: %s, Path: %s\n", c.Request.Method, c.Request.URL.Path)
		fmt.Printf("Authorization header: %s\n", c.GetHeader("Authorization"))

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS request - CORS preflight, returning 204")
			c.AbortWithStatus(204)
			return
		}

		fmt.Println("Not OPTIONS request - continuing to next middleware")
		c.Next()
	})

	// Trust all proxies
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.NewRepository(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	// Инициализация JWT сервиса из конфигурации
	jwtService := service.NewJWTService(conf.JWT.Token, conf.JWT.ExpiresIn)

	// Настройка маршрутов с аутентификацией
	setupRoutes(router, hand, jwtService)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}

// Остальной код без изменений...
func setupRoutes(router *gin.Engine, hand *handler.Handler, jwtService *service.JWTService) {
	// Загрузка HTML шаблонов и статических файлов
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "resources/styles")
	router.Static("/images", "resources/images")

	// Swagger документация
	router.Static("/swagger", "./docs")

	// HTML маршруты (публичные)
	router.GET("/", hand.GetRumbs)
	router.GET("/rumb/:id", hand.GetRumbPage)
	router.GET("/fly_calculation/:id", hand.GetFlyRequestPage)
	router.GET("/deleteRequest/:id", hand.DeleteFlyRequest)

	// API маршруты с префиксом /api
	api := router.Group("/api")
	{
		// Публичные API маршруты (не требуют аутентификации)
		public := api.Group("")
		{
			public.POST("/auth/register", hand.RegisterUser)
			public.POST("/auth/login", hand.LoginUser)
			public.GET("/rumbs", hand.GetRumbsAPI)
			public.GET("/rumbs/:id", hand.GetRumbAPI)
			public.GET("/rumbs/addToRequest/:id", hand.AddToRequest)
		}

		// Защищенные маршруты (требуют аутентификации)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService)) // <- JWT middleware применяется здесь
		{
			fmt.Println("Setting up protected routes with JWT middleware")

			// Пользовательские endpoints
			protected.GET("/auth/me", hand.GetCurrentUser)
			protected.GET("/auth/validate", hand.ValidateToken) // Добавьте этот эндпоинт для тестирования
			protected.PUT("/auth/me", hand.UpdateCurrentUser)
			protected.POST("/auth/logout", hand.LogoutUser)

			// Заявки пользователя
			protected.GET("/flyRequests", hand.GetFlyRequestsAPI)
			protected.GET("/flyRequests/current", hand.GetCurrentFlyRequest)
			protected.POST("/flyRequests", hand.CreateFlyRequest)
			protected.GET("/flyRequests/:id", hand.GetFlyRequestAPI)
			protected.PUT("/flyRequests/:id", hand.UpdateFlyRequest)
			protected.PUT("/flyRequests/:id/form", hand.FormRequest)

			// Работа с румбами в заявках
			protected.POST("/rumbs/addToRequest/:id", hand.AddToRequest)
			protected.DELETE("/flyRequests/rumbs/:flyRequestID/:rumbID", hand.DeleteFlyRequestRumb)
			protected.PUT("/flyRequests/rumbs/:flyRequestID/:rumbID", hand.UpdateFlyRequestRumb)
		}

		// Маршруты только для модераторов
		moderator := api.Group("")
		moderator.Use(middleware.AuthMiddleware(jwtService))
		moderator.Use(middleware.RequireModerator())
		{
			// Управление румбами (CRUD)
			moderator.POST("/rumbs", hand.CreateRumb)
			moderator.PUT("/rumbs/:id", hand.UpdateRumb)
			moderator.DELETE("/rumbs/:id", hand.DeleteRumb)
			moderator.POST("/rumbs/:id/image", hand.UploadRumbImage)

			// Модерация заявок
			moderator.PUT("/flyRequests/:id/finish", hand.FinishRequest)
			moderator.PUT("/flyRequests/:id/calculatedBy", hand.UpdateFlyRequestCalculatedBy)
			moderator.DELETE("/flyRequests/:id", hand.DeleteFlyRequest)
		}
	}
}
