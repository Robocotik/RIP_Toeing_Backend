package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetFlyRequestPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	fly_request, err := h.Repository.GetFlyRequestByID(id)
	if err != nil {
		logrus.Error(err)
	}


	// Если заявка в статусе deleted — перекидываем на /
	if fly_request.Status == "deleted" {
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	
	requestInfo, err := h.Repository.GetCurrentRequestInfo(1)
	if err != nil {
		logrus.Error("Failed to get current request info:", err)
		requestInfo = repository.CurrentRequestInfo{
			RequestID: 1,
			RumbCount: 0,
		}
	}

	rumbs, err := h.Repository.GetRumbsByFlyRequestID(requestInfo.RequestID)
	
	if err != nil {
		logrus.Error("Failed to get rumbs from request: ", err)
	}

	ctx.HTML(http.StatusOK, "fly_calculation.html", gin.H{
		"request":    fly_request,
		"rumbs":      rumbs,
		"request_id": requestInfo.RequestID,
		"rumb_count": requestInfo.RumbCount,
	})

}

// === API для FlyRequest ===

func (h *Handler) GetFlyRequestsAPI(ctx *gin.Context) {
	reqs, err := h.Repository.GetFlyRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}

func (h *Handler) GetFlyRequestAPI(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	req, err := h.Repository.GetFlyRequestByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, req)
}

func (h *Handler) CreateFlyRequest(ctx *gin.Context) {
	var req ds.FlyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if err := h.Repository.CreateFlyRequest(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, req)
}

func (h *Handler) UpdateFlyRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var req ds.FlyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	req.ID = id
	if err := h.Repository.UpdateFlyRequest(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteFlyRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteFlyRequest(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) AddToRequest(ctx *gin.Context) {
	idStr := ctx.Param("id") // <-- id из пути /addToRequest/:id
	rumbID, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid ID parameter:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid rumb id"})
		return
	}

	// Проверяем, что румб существует
	_, err = h.Repository.GetRumbByID(rumbID)
	if err != nil {
		logrus.Error("Rumb not found:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rumb not found"})
		return
	}

	// Получаем или создаём активную заявку
	flyRequest, err := h.Repository.GetFlyRequestByStatus("created")
	if err != nil {
		newFlyRequest := &ds.FlyRequest{
			Status:      "created",
			CreatedByID: 1,
			ModeratorID: 1,
		}
		if err := h.Repository.CreateFlyRequest(newFlyRequest); err != nil {
			logrus.Error("Failed to create fly request:", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		flyRequest = *newFlyRequest
	}

	// Находим максимальный SegmentOrder
	maxSegmentOrder, err := h.Repository.GetMaxSegmentOrder(flyRequest.ID)
	if err != nil {
		logrus.Error("Failed to get max segment order:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Добавляем связь FlyRequest_Rumb
	flyRequestRumb := ds.FlyRequest_Rumb{
		FlyRequestID: uint(flyRequest.ID),
		RumbID:       uint(rumbID),
		SegmentOrder: maxSegmentOrder + 1,
		DistanceKM:   100.0,
		WindSpeedKMH: 50.0,
		IsMain:       false,
	}

	if err := h.Repository.CreateFlyRequestRumb(flyRequestRumb); err != nil {
		logrus.Error("Failed to create fly request rumb link:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to request"})
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) DeleteFlyRequestRumb(ctx *gin.Context) {
	flyRequestID, err := strconv.Atoi(ctx.Param("flyRequestID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid fly request id"})
		return
	}
	rumbID, err := strconv.Atoi(ctx.Param("rumbID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid rumb id"})
		return
	}

	err = h.Repository.DeleteFlyRequestRumb(uint(flyRequestID), uint(rumbID))
	if err != nil {
		logrus.Error("Failed to delete fly request rumb:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete rumb from request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "rumb deleted from fly request"})
}

func (h *Handler) UpdateFlyRequestRumb(ctx *gin.Context) {
	flyRequestID, err := strconv.Atoi(ctx.Param("flyRequestID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid fly request id"})
		return
	}
	rumbID, err := strconv.Atoi(ctx.Param("rumbID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid rumb id"})
		return
	}

	var input struct {
		DistanceKM   float64 `json:"distance_km"`
		WindSpeedKMH float64 `json:"wind_speed_kmh"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err = h.Repository.UpdateFlyRequestRumb(uint(flyRequestID), uint(rumbID), input.DistanceKM, input.WindSpeedKMH)
	if err != nil {
		logrus.Error("Failed to update fly request rumb:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update rumb in request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"fly_request_id": flyRequestID,
		"rumb_id":        rumbID,
		"distance_km":    input.DistanceKM,
		"wind_speed_kmh": input.WindSpeedKMH,
	})
}

// func (h *Handler) DeleteFlyRequest(ctx *gin.Context) {
// 	idStr := ctx.Param("id")
// 	flyRequestID, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		logrus.Error("Invalid ID parameter:", err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
// 		return
// 	}

	
// 	_, err = h.Repository.GetFlyRequestByID(flyRequestID)
// 	if err != nil {
// 		logrus.Error("FlyRequest not found:", err)
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "FlyRequest not found"})
// 		return
// 	}

	
// 	err = h.Repository.UpdateFlyRequestStatus(flyRequestID, "deleted")
// 	if err != nil {
// 		logrus.Error("Failed to update fly request status:", err)
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete request"})
// 		return
// 	}

// 	ctx.Redirect(http.StatusFound, "/")
// }
