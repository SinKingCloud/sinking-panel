package oss

import "server/app/util/oss/types"

// 对外保留统一类型入口，厂商子包共同依赖 types，避免循环依赖。
type (
	Config          = types.Config
	Bucket          = types.Bucket
	Object          = types.Object
	ListOptions     = types.ListOptions
	ObjectPage      = types.ObjectPage
	PutOptions      = types.PutOptions
	GetOptions      = types.GetOptions
	Download        = types.Download
	CopyOptions     = types.CopyOptions
	CopyPartOptions = types.CopyPartOptions
	Part            = types.Part
	MultipartUpload = types.MultipartUpload
	PartsOptions    = types.PartsOptions
	PartsPage       = types.PartsPage
	UploadsOptions  = types.UploadsOptions
	UploadsPage     = types.UploadsPage
	SignedURL       = types.SignedURL
	DeleteResult    = types.DeleteResult
	UploadOptions   = types.UploadOptions
)

// Manager 绑定一个桶，可并发调用；所有网络操作都接受 context。
type Manager struct {
	config Config
	driver backend
}
