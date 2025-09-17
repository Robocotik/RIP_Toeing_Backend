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

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		rumbs, err = h.Repository.GetRumbs()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		rumbs, err = h.Repository.GetRumbsByTitle(searchQuery) // в ином случае ищем заказ по заголовку
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
		"query": searchQuery, // передаем введенный запрос обратно на страницу
		"request_id": requestInfo.RequestID,
		"rumb_count": requestInfo.RumbCount,
		// в ином случае оно будет очищаться при нажатии на кнопку
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
