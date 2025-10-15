package service

import (
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
	"path/filepath"
)

// SaveUploadedImage 保存上传的文件并返回URL
func SaveUploadedImage(ctx *gin.Context, file *multipart.FileHeader) (string, error) {
	savePath := "./static/uploads/" + filepath.Base(file.Filename)

	// 创建目录
	if err := os.MkdirAll("./static/uploads", os.ModePerm); err != nil {
		return "", err
	}

	// 保存文件
	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	// 返回相对路径
	return "/static/uploads/" + file.Filename, nil
}
