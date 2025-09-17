package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/repository"
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
		flyRequest = *newFlyRequest
	}

	// Получаем текущий максимальный segment_order для этой заявки
	maxSegmentOrder, err := h.Repository.GetMaxSegmentOrder(flyRequest.ID)
	if err != nil {
		logrus.Error("Failed to get max segment order:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Создаем связь между FlyRequest и Rumb используя существующую структуру
	flyRequestRumb := ds.FlyRequest_Rumb{
		FlyRequestID: uint(flyRequest.ID),
		RumbID:       uint(rumbID),
		SegmentOrder: maxSegmentOrder + 1,
		DistanceKM:   100.0,
		WindSpeedKMH: 50.0,
		IsMain:       false,
	}

	err = h.Repository.CreateFlyRequestRumb(flyRequestRumb)
	if err != nil {
		logrus.Error("Failed to create fly request rumb link:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to request"})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) GetFlyRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	fly_request, err := h.Repository.GetFlyRequest()
	if err != nil {
		logrus.Error(err)
	}

	// Получаем информацию о текущей заявке
	requestInfo, err := h.Repository.GetCurrentRequestInfo(1)
	if err != nil {
		logrus.Error("Failed to get current request info:", err)
		requestInfo = repository.CurrentRequestInfo{
			RequestID: 1,
			RumbCount: 0,
		}
	}

	ctx.HTML(http.StatusOK, "fly_calculation.html", gin.H{
		"request":    fly_request,
		"request_id": requestInfo.RequestID,
		"rumb_count": requestInfo.RumbCount,
	})
}
