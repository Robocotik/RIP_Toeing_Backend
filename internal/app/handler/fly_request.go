package handler

import (
	"backend/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) AddToRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	rumbID, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid ID parameter:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Проверяем существование румба
	_, err = h.Repository.GetRumb(rumbID)
	if err != nil {
		logrus.Error("Rumb not found:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rumb not found"})
		return
	}

	var flyRequest ds.FlyRequest
	
	// Ищем существующую заявку со статусом "created"
	flyRequest, err = h.Repository.GetFlyRequestByStatus("created")
	if err != nil {
		// Если нет заявки со статусом "created", создаем новую
		newFlyRequest := &ds.FlyRequest{
			Status:      "created",
			CreatedByID: 1,
			ModeratorID: 1,
		}
		err = h.Repository.CreateFlyRequest(newFlyRequest)
		if err != nil {
			logrus.Error("Failed to create fly request:", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		flyRequest = *newFlyRequest // Используем созданную заявку с заполненным ID
	}

	// Получаем текущий максимальный segment_order для этой заявки
	maxSegmentOrder, err := h.Repository.GetMaxSegmentOrder(flyRequest.ID)
	if err != nil {
		logrus.Error("Failed to get max segment order:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Создаем связь между FlyRequest и Rumb
	flyRequestRumb := ds.FlyRequest_Rumb{
		FlyRequestID: uint(flyRequest.ID),
		RumbID:       uint(rumbID),
		SegmentOrder: maxSegmentOrder + 1,
		DistanceKM:   100.0,    // Примерное значение, нужно получить из данных
		WindSpeedKMH: 50.0,     // Примерное значение, нужно получить из данных
		IsMain:       false,    // или true в зависимости от логики
	}

	err = h.Repository.CreateFlyRequestRumb(flyRequestRumb)
	if err != nil {
		logrus.Error("Failed to create fly request rumb link:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to request"})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}