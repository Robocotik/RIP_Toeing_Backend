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
	// Читаем query-параметры
	status := ctx.Query("status")             // статус для фильтрации, если передан
	formedAfter := ctx.Query("formedAfter")   // фильтр по FormedAt >= переданной дате
	formedBefore := ctx.Query("formedBefore") // фильтр по FormedAt <= переданной дате

	reqs, err := h.Repository.GetFlyRequests(status, formedAfter, formedBefore)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Формируем массив для отдачи, исключая ModeratorID и CreatedByID
	resp := make([]gin.H, 0, len(reqs))
	for _, r := range reqs {
		resp = append(resp, gin.H{
			"ID":           r.ID,
			"Status":       r.Status,
			"CreatedAt":    r.CreatedAt,
			"FormedAt":     r.FormedAt,
			"CalculatedBy": r.CalculatedBy,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GetFlyRequestAPI(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid fly request id"})
		return
	}

	// Получаем заявку
	req, err := h.Repository.GetFlyRequestByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}

	// Получаем все румбы заявки
	rumbLinks, err := h.Repository.GetRumbsByFlyRequestID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении услуг"})
		return
	}

	// Формируем массив с нужными полями румба
	rumbs := make([]gin.H, 0, len(rumbLinks))
	for _, link := range rumbLinks {
		rumbs = append(rumbs, gin.H{
			"id":           link.Rumb.ID,
			"title":        link.Rumb.Title,
			"image":        link.Rumb.Image,
			"SegmentOrder": link.SegmentOrder,
			"DistanceKM":   link.DistanceKM,
			"WindSpeedKMH": link.WindSpeedKMH,
			"IsMain":       link.IsMain,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"fly_request": req,
		"rumbs":       rumbs,
	})
}

func (h *Handler) CreateFlyRequest(ctx *gin.Context) {
	req := ds.FlyRequest{
		Status:      "draft",
		CreatedByID: 1,
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
	idStr := ctx.Param("id") // id румба
	rumbID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid rumb id"})
		return
	}

	// Проверяем, что румб существует
	_, err = h.Repository.GetRumbByID(rumbID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rumb not found"})
		return
	}

	// Получаем активную заявку draft
	flyRequest, err := h.Repository.GetFlyRequestByStatus("draft")
	if err != nil || flyRequest.ID == 0 {
		// Если заявки нет — создаём через CreateFlyRequest
		h.CreateFlyRequest(ctx)
		// Получаем только что созданную заявку
		flyRequest, err = h.Repository.GetFlyRequestByStatus("draft")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch newly created request"})
			return
		}
	}

	// Находим максимальный SegmentOrder
	maxSegmentOrder, err := h.Repository.GetMaxSegmentOrder(flyRequest.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get max segment order"})
		return
	}

	// Добавляем румб в заявку
	flyRequestRumb := ds.FlyRequest_Rumb{
		FlyRequestID: uint(flyRequest.ID),
		RumbID:       uint(rumbID),
		SegmentOrder: maxSegmentOrder + 1,
		DistanceKM:   100.0,
		WindSpeedKMH: 50.0,
		IsMain:       false,
	}

	if err := h.Repository.CreateFlyRequestRumb(flyRequestRumb); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add rumb to request"})
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

func (h *Handler) UpdateFlyRequestCalculatedBy(ctx *gin.Context) {
	// Получаем ID заявки
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid fly request id"})
		return
	}

	// Парсим тело запроса
	var input struct {
		CalculatedBy string `json:"calculated_by"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Получаем заявку
	req, err := h.Repository.GetFlyRequestByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "fly request not found"})
		return
	}

	// Обновляем поле CalculatedBy
	req.CalculatedBy = input.CalculatedBy
	if err := h.Repository.UpdateFlyRequest(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update fly request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            req.ID,
		"calculated_by": req.CalculatedBy,
	})
}

func (h *Handler) FormRequest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, rumbs, err := h.Repository.GetByIDWithRumbs(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	// Проверяем каждый сегмент
	for _, seg := range rumbs {
		if seg.DistanceKM == 0 || seg.WindSpeedKMH == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All segments must have DistanceKM and WindSpeedKMH"})
			return
		}
	}


	if err := h.Repository.UpdateStatus(uint(id), "formed", nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request formed successfully"})
}


func (h *Handler) FinishRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var body struct {
		Action      string `json:"action"`      // "complete" или "reject"
		ModeratorID int    `json:"moderatorID"` // ID модератора
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	var newStatus string
	switch body.Action {
	case "complete":
		newStatus = "finished"
	case "reject":
		newStatus = "rejected"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action, must be 'complete' or 'reject'"})
		return
	}

	if err := h.Repository.UpdateStatus(uint(id), newStatus, &body.ModeratorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Request status updated successfully",
		"new_status":  newStatus,
		"moderatorID": body.ModeratorID,
	})
}


func (h *Handler) GetCurrentFlyRequest(c *gin.Context) {
    userID := 1 // жестко для примера, обычно берется из контекста

    fr, count, err := h.Repository.GetDraftByUser(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch fly request"})
        return
    }

    if fr == nil {
        c.JSON(http.StatusOK, gin.H{"flyRequestID": nil, "rumbsCount": 0})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "flyRequestID":  fr.ID,
        "rumbsCount": count,
    })
}