package handler

import (

	"backend/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Загружаем HTML-шаблоны и стили
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "resources/styles")
	router.Static("/images", "resources/images")

	// === СТАРЫЕ СТРАНИЦЫ ===
	router.GET("/", h.GetRumbs)
	router.GET("/rumb/:id", h.GetRumbPage)
	router.GET("/fly_calculation/:id", h.GetFlyRequestPage)
	router.GET("/addToRequest/:id", h.AddToRequest)
	router.GET("/deleteRequest/:id", h.DeleteFlyRequest)

	// === НОВЫЙ API ===
	api := router.Group("/api")
	{
		// Rumbs
		api.GET("/rumbs", h.GetRumbsAPI)
		api.GET("/rumbs/:id", h.GetRumbAPI)
		api.POST("/rumbs", h.CreateRumb)
		api.PUT("/rumbs/:id", h.UpdateRumb)
		api.DELETE("/rumbs/:id", h.DeleteRumb)

		// Fly Requests
		api.GET("/flyRequests", h.GetFlyRequestsAPI)
		api.GET("/flyRequests/:id", h.GetFlyRequestAPI)
		api.POST("/flyRequests", h.CreateFlyRequest)
		api.PUT("/flyRequests/:id", h.UpdateFlyRequest)
		api.DELETE("/flyRequests/:id", h.DeleteFlyRequest)

		// Auth
		api.POST("/auth/register", h.RegisterUser)
		api.POST("/auth/login", h.LoginUser)
	}
}