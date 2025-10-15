package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetFlyRequestPage godoc
// @Summary Получить страницу заявки на полет
// @Description Отображает HTML страницу с детальной информацией о заявке на полет
// @Tags Frontend
// @Accept html
// @Produce html
// @Param id path int true "ID заявки на полет"
// @Success 200 {string} string "HTML страница заявки"
// @Failure 400 {string} string "Неверный ID заявки"
// @Failure 404 {string} string "Заявка не найдена"
// @Router /fly_calculation/{id} [get]
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

// GetFlyRequestsAPI godoc
// @Summary Получить список заявок на полет
// @Description Возвращает список заявок на полет с возможностью фильтрации по статусу и датам
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param status query string false "Фильтр по статусу (draft, formed, finished, rejected)"
// @Param formedAfter query string false "Фильтр по дате формирования (>=) в формате YYYY-MM-DD"
// @Param formedBefore query string false "Фильтр по дате формирования (<=) в формате YYYY-MM-DD"
// @Success 200 {array} object "Список заявок"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests [get]
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

// GetFlyRequestAPI godoc
// @Summary Получить детальную информацию о заявке на полет
// @Description Возвращает полную информацию о заявке на полет включая связанные румбы
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Success 200 {object} object "Детальная информация о заявке"
// @Failure 400 {object} object "Неверный ID заявки"
// @Failure 404 {object} object "Заявка не найдена"
// @Router /api/flyRequests/{id} [get]
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

// CreateFlyRequest godoc
// @Summary Создать новую заявку на полет
// @Description Создает новую заявку на полет в статусе "draft"
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Success 201 {object} ds.FlyRequest "Созданная заявка"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests [post]
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

// UpdateFlyRequest godoc
// @Summary Обновить заявку на полет
// @Description Обновляет информацию о заявке на полет
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Param request body ds.FlyRequest true "Данные для обновления заявки"
// @Success 200 {object} ds.FlyRequest "Обновленная заявка"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/{id} [put]
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

// DeleteFlyRequest godoc
// @Summary Удалить заявку на полет
// @Description Удаляет заявку на полет по ID
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Success 302 "Перенаправление на главную страницу"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/{id} [delete]
func (h *Handler) DeleteFlyRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteFlyRequest(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

// AddToRequest godoc
// @Summary Добавить румб в заявку
// @Description Добавляет указанный румб в активную заявку или создает новую заявку
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID румба для добавления"
// @Success 302 "Перенаправление на главную страницу"
// @Failure 400 {object} object "Неверный ID румба"
// @Failure 404 {object} object "Румб не найден"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/rumbs/addToRequest/{id} [post]
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

// DeleteFlyRequestRumb godoc
// @Summary Удалить румб из заявки
// @Description Удаляет связь между румбом и заявкой на полет
// @Tags FlyRequestRumbs
// @Accept json
// @Produce json
// @Param flyRequestID path int true "ID заявки на полет"
// @Param rumbID path int true "ID румба"
// @Success 200 {object} object "Сообщение об успешном удалении"
// @Failure 400 {object} object "Неверные ID"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/rumbs/{flyRequestID}/{rumbID} [delete]
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

// UpdateFlyRequestRumb godoc
// @Summary Обновить параметры румба в заявке
// @Description Обновляет расстояние и скорость ветра для румба в заявке
// @Tags FlyRequestRumbs
// @Accept json
// @Produce json
// @Param flyRequestID path int true "ID заявки на полет"
// @Param rumbID path int true "ID румба"
// @Param request body object true "Данные для обновления" { "distance_km": 150.5, "wind_speed_kmh": 25.3 }
// @Success 200 {object} object "Обновленные данные"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/rumbs/{flyRequestID}/{rumbID} [put]
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

// UpdateFlyRequestCalculatedBy godoc
// @Summary Обновить поле CalculatedBy заявки
// @Description Обновляет информацию о том, кем был выполнен расчет заявки
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Param request body object true "Данные для обновления" { "calculated_by": "auto_calc_system" }
// @Success 200 {object} object "Обновленная информация"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 404 {object} object "Заявка не найдена"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/{id}/calculatedBy [put]
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

// FormRequest godoc
// @Summary Сформировать заявку
// @Description Переводит заявку в статус "formed" после проверки всех сегментов
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Success 200 {object} object "Сообщение об успешном формировании"
// @Failure 400 {object} object "Не все сегменты заполнены"
// @Failure 404 {object} object "Заявка не найдена"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/{id}/form [put]
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

// FinishRequest godoc
// @Summary Завершить или отклонить заявку
// @Description Переводит заявку в статус "finished" или "rejected" с указанием модератора
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки на полет"
// @Param request body object true "Данные для завершения" { "action": "complete", "moderatorID": 456 }
// @Success 200 {object} object "Результат операции"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/{id}/finish [put]
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

// GetCurrentFlyRequest godoc
// @Summary Получить текущую заявку пользователя
// @Description Возвращает ID текущей черновой заявки и количество румбов в ней
// @Tags FlyRequests
// @Accept json
// @Produce json
// @Success 200 {object} object "Информация о текущей заявке"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/flyRequests/current [get]
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