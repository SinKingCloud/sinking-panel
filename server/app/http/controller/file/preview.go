package file

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"strconv"
)

func Preview(c *context.Context) {
	var form struct {
		Path     string `json:"path" default:"" validate:"required" label:"文件路径"`
		Download bool   `json:"download" default:"false" validate:"omitempty" label:"是否下载"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.Path) {
		c.Error("文件不存在")
		return
	}
	if f.IsDir(form.Path) {
		c.Error("不能预览目录")
		return
	}
	// 获取文件信息
	fileInfo, err := f.FileInfo(form.Path)
	if err != nil {
		c.Error("获取文件信息失败: " + err.Error())
		return
	}
	// 获取文件路径
	filePath := f.Path(form.Path, true, true)
	// 打开文件
	f2, err := os.Open(filePath)
	if err != nil {
		c.Error("打开文件失败: " + err.Error())
		return
	}
	defer f2.Close()
	// 获取文件名并进行URL编码，确保特殊字符和中文能正确显示
	fileName := fileInfo.Name
	encodedFileName := url.QueryEscape(fileName)
	// 设置文件类型
	contentType := service.File.GetContentType(fileName)
	// 设置响应头
	c.Writer.Header().Set("Content-Type", contentType)
	// 如果是下载模式，添加下载头
	if form.Download {
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"; filename*=UTF-8''"+encodedFileName)
	} else {
		// 对于某些文件类型，如PDF、图片等，直接在浏览器中查看
		if service.File.IsViewableInBrowser(contentType) {
			c.Writer.Header().Set("Content-Disposition", "inline; filename=\""+fileName+"\"; filename*=UTF-8''"+encodedFileName)
		} else {
			c.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"; filename*=UTF-8''"+encodedFileName)
		}
	}
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	c.SetStatus(http.StatusOK)
	_, _ = io.Copy(c.Writer, f2)
}
