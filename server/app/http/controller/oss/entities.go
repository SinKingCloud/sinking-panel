package oss

import utilOss "server/app/util/oss"

// fileOptions 使用与 HTTP 接口一致的字段名描述文件属性。
type fileOptions struct {
	ContentType        string            `json:"content_type" validate:"omitempty,max=256"`
	CacheControl       string            `json:"cache_control" validate:"omitempty,max=1024"`
	ContentDisposition string            `json:"content_disposition" validate:"omitempty,max=1024"`
	ContentEncoding    string            `json:"content_encoding" validate:"omitempty,max=256"`
	StorageClass       string            `json:"storage_class" validate:"omitempty,max=100"`
	Metadata           map[string]string `json:"metadata" validate:"omitempty,max=100"`
}

// putOptions 转为各厂商共用的文件属性。
func (o fileOptions) putOptions() utilOss.PutOptions {
	return utilOss.PutOptions{
		ContentType: o.ContentType, CacheControl: o.CacheControl,
		ContentDisposition: o.ContentDisposition, ContentEncoding: o.ContentEncoding,
		StorageClass: o.StorageClass, Metadata: o.Metadata,
	}
}

// deleteResult 返回逐个文件的删除结果，保留失败文件及原因。
type deleteResult struct {
	Key   string `json:"key"`
	Error string `json:"error,omitempty"`
}
