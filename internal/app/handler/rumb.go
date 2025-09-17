package handler

import (
	"backend/internal/app/ds"
	"backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		"rumbs": rumbs,
		"query": searchQuery, 
		"request_id": requestInfo.RequestID,
		"rumb_count": requestInfo.RumbCount,
		
	})
}

func (h *Handler) GetRumb(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	rumb, err := h.Repository.GetRumb(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "rumb.html", gin.H{
		"rumb": rumb,
	})
}


// GetRumbsByFlyRequestIDFull - метод для получения полных данных румб по ID заявки
func (h *Handler) GetRumbsByFlyRequestIDFull(ctx *gin.Context) {
	idStr := ctx.Param("id")
	flyRequestID, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid ID parameter:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Получаем румбы для указанной заявки
	rumbs, err := h.Repository.GetRumbsByFlyRequestID(flyRequestID)
	if err != nil {
		logrus.Error("Failed to get rumbs for fly request:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rumbs"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"fly_request_id": flyRequestID,
		"rumbs":          rumbs,
		"count":          len(rumbs),
	})
}