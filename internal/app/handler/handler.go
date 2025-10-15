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

	router.GET("/", h.GetRumbs)
	router.GET("/rumb/:id", h.GetRumbPage)
	router.GET("/fly_calculation/:id", h.GetFlyRequestPage)
	router.GET("/deleteRequest/:id", h.DeleteFlyRequest)
	router.GET("/api/rumbs/addToRequest/:id", h.AddToRequest)

	api := router.Group("/api")
	{
		// Rumbs
		api.GET("/rumbs", h.GetRumbsAPI)
		api.GET("/rumbs/:id", h.GetRumbAPI)
		api.POST("/rumbs", h.CreateRumb)
		api.PUT("/rumbs/:id", h.UpdateRumb)
		api.DELETE("/rumbs/:id", h.DeleteRumb)
		api.POST("/rumbs/addToRequest/:id", h.AddToRequest)
		api.POST("/rumbs/:id/image", h.UploadRumbImage)
		

		// Fly Requests
		api.GET("/flyRequests", h.GetFlyRequestsAPI)
		api.GET("/flyRequests/:id", h.GetFlyRequestAPI)
		api.POST("/flyRequests", h.CreateFlyRequest)
		api.PUT("/flyRequests/:id", h.UpdateFlyRequest)
		api.DELETE("/flyRequests/:id", h.DeleteFlyRequest)
		api.PUT("/flyRequests/:id/calculatedBy", h.UpdateFlyRequestCalculatedBy)
		api.PUT("/flyRequests/:id/form", h.FormRequest)
		api.PUT("/flyRequests/:id/finish", h.FinishRequest)
		api.GET("/flyRequests/current", h.GetCurrentFlyRequest)

		// Fly Request Rumbs
		api.DELETE("/flyRequests/rumbs/:flyRequestID/:rumbID", h.DeleteFlyRequestRumb)
		api.PUT("/flyRequests/rumbs/:flyRequestID/:rumbID", h.UpdateFlyRequestRumb)

		// Auth
		api.POST("/auth/register", h.RegisterUser)
		api.POST("/auth/login", h.LoginUser)
		api.POST("/auth/logout", h.LogoutUser)
		api.GET("/auth/me", h.GetCurrentUser)
		api.PUT("/auth/me", h.UpdateCurrentUser)

	}
}
