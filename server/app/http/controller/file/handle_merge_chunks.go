package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

func handleMergeChunks(c *context.Context) {
	meta, err := parseUploadMeta(
		c.DefaultForm("upload_id", ""),
		c.DefaultForm("path", "/"),
		c.DefaultForm("file_name", ""),
		c.DefaultForm("total_size", ""),
		c.DefaultForm("chunk_size", ""),
		c.DefaultForm("total_chunks", ""),
		c.DefaultForm("file_hash", ""),
	)
	if err != nil {
		c.Error(err.Error())
		return
	}
	completed, existed, err := service.File.MergeUpload(c.Request.Context(), meta)
	if err != nil {
		c.Error(err.Error())
	} else {
		if !existed {
			service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "合并上传文件", "上传文件["+completed.Path+"]")
		}
		c.SuccessWithData("文件合并成功", completed)
	}
}
