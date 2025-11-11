package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// === СТАРЫЕ РУЧКИ — отрисовка страниц ===

// GetRumbs godoc
// @Summary Главная страница с румбами
// @Description Отображает HTML страницу со списком румбов с возможностью поиска и информацией о текущей заявке
// @Tags Frontend
// @Accept html
// @Produce html
// @Param query query string false "Поисковый запрос для фильтрации румбов по названию"
// @Success 200 {string} string "HTML страница с румбами"
// @Router / [get]
func (h *Handler) GetRumbs(ctx *gin.Context) {
	var rumbs []ds.Rumb
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		rumbs, err = h.Repository.GetRumbs()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		rumbs, err = h.Repository.GetRumbsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	requestInfo, err := h.Repository.GetCurrentRequestInfo(1)
	if err != nil {
		logrus.Error("Failed to get current request info:", err)
		requestInfo = repository.CurrentRequestInfo{
			RequestID: 1,
			RumbCount: 0,
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"rumbs":      rumbs,
		"query":      searchQuery,
		"request_id": requestInfo.RequestID,
		"rumb_count": requestInfo.RumbCount,
	})
}

// GetRumbPage godoc
// @Summary Страница детальной информации о румбе
// @Description Отображает HTML страницу с детальной информацией о конкретном румбе
// @Tags Frontend
// @Accept html
// @Produce html
// @Param id path int true "ID румба"
// @Success 200 {string} string "HTML страница румба"
// @Failure 400 {string} string "Неверный ID румба"
// @Failure 404 {string} string "Румб не найден"
// @Router /rumb/{id} [get]
func (h *Handler) GetRumbPage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Неверный ID"})
		return
	}

	rumb, err := h.Repository.GetRumbByID(id)
	if err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Румб не найден"})
		return
	}

	ctx.HTML(http.StatusOK, "rumb.html", gin.H{
		"rumb": rumb,
	})
}

// GetRumbsAPI godoc
// @Summary Получить список всех румбов
// @Description Возвращает JSON массив всех румбов в системе с возможностью фильтрации по названию
// @Tags Rumbs
// @Accept json
// @Produce json
// @Param query query string false "Поисковый запрос для фильтрации румбов по названию"
// @Success 200 {array} ds.Rumb "Список румбов"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/rumbs [get]
func (h *Handler) GetRumbsAPI(ctx *gin.Context) {
	var rumbs []ds.Rumb
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		rumbs, err = h.Repository.GetRumbs()
	} else {
		rumbs, err = h.Repository.GetRumbsByTitle(searchQuery)
	}

	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rumbs)
}

// GetRumbAPI godoc
// @Summary Получить информацию о румбе по ID
// @Description Возвращает детальную информацию о конкретном румбе
// @Tags Rumbs
// @Accept json
// @Produce json
// @Param id path int true "ID румба"
// @Success 200 {object} ds.Rumb "Информация о румбе"
// @Failure 400 {object} object "Неверный ID румба"
// @Failure 404 {object} object "Румб не найден"
// @Router /api/rumbs/{id} [get]
func (h *Handler) GetRumbAPI(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	rumb, err := h.Repository.GetRumbByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Румб не найден"})
		return
	}
	ctx.JSON(http.StatusOK, rumb)
}

// CreateRumb godoc
// @Summary Создать новый румб
// @Description Создает новый румб в системе
// @Tags Rumbs
// @Accept json
// @Produce json
// @Param rumb body ds.Rumb true "Данные для создания румба"
// @Success 201 {object} ds.Rumb "Созданный румб"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/rumbs [post]
func (h *Handler) CreateRumb(ctx *gin.Context) {
	var rumb ds.Rumb
	if err := ctx.ShouldBindJSON(&rumb); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	rumb.ID = 0 // важно! чтобы не конфликтовало с PK
	if err := h.Repository.CreateRumb(&rumb); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rumb)
}

// UpdateRumb godoc
// @Summary Обновить информацию о румбе
// @Description Обновляет информацию о существующем румбе
// @Tags Rumbs
// @Accept json
// @Produce json
// @Param id path int true "ID румба"
// @Param rumb body ds.Rumb true "Новые данные румба"
// @Success 200 {object} ds.Rumb "Обновленный румб"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/rumbs/{id} [put]
func (h *Handler) UpdateRumb(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var rumb ds.Rumb
	if err := ctx.ShouldBindJSON(&rumb); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	rumb.ID = id
	if err := h.Repository.UpdateRumb(&rumb); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rumb)
}

// DeleteRumb godoc
// @Summary Удалить румб
// @Description Удаляет румб из системы по ID
// @Tags Rumbs
// @Accept json
// @Produce json
// @Param id path int true "ID румба"
// @Success 200 {object} object "Сообщение об успешном удалении"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/rumbs/{id} [delete]
func (h *Handler) DeleteRumb(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteRumb(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
