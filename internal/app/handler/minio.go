package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UploadRumbImage godoc
// @Summary Загрузить изображение для румба
// @Description Загружает изображение для указанного румба в MinIO
// @Tags Rumbs
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID румба"
// @Param image formData file true "Изображение для загрузки"
// @Success 200 {object} object "Информация об обновленном румбе"
// @Failure 400 {object} object "Неверные данные запроса"
// @Failure 404 {object} object "Румб не найден"
// @Failure 500 {object} object "Ошибка загрузки изображения"
// @Router /api/rumbs/{id}/image [post]
func (h *Handler) UploadRumbImage(ctx *gin.Context) {
	// Получаем ID услуги из URL
	idStr := ctx.Param("id")
	rumbID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid rumb id"})
		return
	}

	// Проверяем, что услуга существует
	rumb, err := h.Repository.GetRumbByID(rumbID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "rumb not found"})
		return
	}

	// Получаем файл из form-data
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	// Используем оригинальное имя файла
	fileName := file.Filename

	// Загружаем в MinIO
	err = h.Repository.UploadRumbImageToMinio(rumbID, file, fileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
		return
	}

	// Обновляем запись в БД
	rumb.Image = fileName
	if err := h.Repository.UpdateRumb(rumb); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update rumb record"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":    rumb.ID,
		"title": rumb.Title,
		"image": rumb.Image,
	})
}