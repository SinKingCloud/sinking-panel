package file_format

const (
	Zip  = "zip"
	Tar  = "tar"
	Gzip = "gz"
	Tgz  = "tgz"
)

// Map 文件压缩格式数据
func Map() map[string]string {
	return map[string]string{
		Zip:  Zip,
		Tar:  Tar,
		Gzip: Gzip,
		Tgz:  Tgz,
	}
}
