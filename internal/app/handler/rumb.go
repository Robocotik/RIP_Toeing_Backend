package handler

import (
	"backend/internal/app/ds"
	"net/http"
	"backend/internal/app/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)


// === СТАРЫЕ РУЧКИ — отрисовка страниц ===

// Главная страница
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

// Страница румба
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

// === API для Rumb ===

func (h *Handler) GetRumbsAPI(ctx *gin.Context) {
	rumbs, err := h.Repository.GetRumbs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rumbs)
}

func (h *Handler) GetRumbAPI(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	rumb, err := h.Repository.GetRumbByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Румб не найден"})
		return
	}
	ctx.JSON(http.StatusOK, rumb)
}

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

func (h *Handler) DeleteRumb(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteRumb(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}