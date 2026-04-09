package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadImage(ctx *gin.Context) {
	// 获取 multipart form 数据
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "表单解析失败: " + err.Error()})
		return
	}

	files := form.File["files"]
	var uploadedURLs []string

	// 确保上传目录存在
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建上传目录: " + err.Error()})
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			// 跳过损坏的文件，继续传下一个
			continue
		}

		// 验证文件真实类型
		buff := make([]byte, 512)
		if _, err := file.Read(buff); err != nil {
			err := file.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			continue
		}

		// 游标复位，非常关键！否则后续保存时图片会丢失前 512 字节导致损坏
		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		contentType := http.DetectContentType(buff)
		if !strings.HasPrefix(contentType, "image/") {
			// 如果不是图片，直接跳过
			continue
		}

		// 生成唯一防冲突文件名 (时间戳 + 原本的后缀)
		ext := filepath.Ext(fileHeader.Filename)
		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		dst := filepath.Join(uploadDir, fileName)

		// 使用 Gin 原生的方式保存，效率更高
		if err := ctx.SaveUploadedFile(fileHeader, dst); err == nil {
			uploadedURLs = append(uploadedURLs, fmt.Sprintf("/uploads/images/%s", fileName))
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "批量上传成功",
		"urls":    uploadedURLs, // 返回一个数组，前端可以直接循环渲染
	})
}
