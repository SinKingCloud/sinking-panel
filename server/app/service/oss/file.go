package oss

import (
	"context"
	"errors"
	"fmt"
	utilOss "server/app/util/oss"
	"slices"
	"strings"
	"unicode/utf8"
)

// Delete 删除文件和目录；目录按前缀分页清空，结果按选中的目标返回。
func (s *service) Delete(ctx context.Context, id int64, keys []string) ([]utilOss.DeleteResult, error) {
	if len(keys) == 0 {
		return nil, errors.New("请选择要删除的文件或目录")
	}
	for _, key := range keys {
		if key == "" || len(key) > 1024 || !utf8.ValidString(key) || strings.ContainsAny(key, "\x00\r\n") {
			return nil, errors.New("文件名称需为 1..1024 字节的有效 UTF-8 字符串且不含控制换行")
		}
	}
	client, err := s.client(id)
	if err != nil {
		return nil, err
	}
	// 排序副本后合并父子目录，避免同一对象因重复选择被多次删除。
	keys = slices.Clone(keys)
	slices.Sort(keys)
	targets := keys[:0]
	for _, key := range keys {
		if len(targets) > 0 {
			previous := targets[len(targets)-1]
			if key == previous || (strings.HasSuffix(previous, "/") && strings.HasPrefix(key, previous)) {
				continue
			}
		}
		targets = append(targets, key)
	}
	results := make([]utilOss.DeleteResult, 0, len(targets))
	var failures []error
	for _, key := range targets {
		if err := ctx.Err(); err != nil {
			return results, errors.Join(append(failures, err)...)
		}
		if strings.HasSuffix(key, "/") {
			options := utilOss.ListOptions{Prefix: key, Limit: 1000}
			for {
				var page *utilOss.FilePage
				page, err = client.ListFiles(ctx, options)
				if err != nil {
					break
				}
				batch := make([]string, 0, len(page.Files))
				for _, file := range page.Files {
					if !strings.HasPrefix(file.Key, key) {
						err = errors.New("云端返回了目录以外的文件，已停止删除")
						break
					}
					if file.Key != key {
						batch = append(batch, file.Key)
					}
				}
				if err == nil && len(batch) > 0 {
					_, err = client.DeleteFiles(ctx, batch)
				}
				if err != nil || !page.HasMore {
					break
				}
				options.Token = page.NextToken
			}
		}
		// 目录标记最后删除，部分失败时保留目录入口，便于重试。
		if err == nil {
			err = client.DeleteFile(ctx, key)
		}
		if err != nil {
			err = fmt.Errorf("删除 %q 失败：%w", key, err)
			failures = append(failures, err)
		}
		results = append(results, utilOss.DeleteResult{Key: key, Err: err})
		err = nil
	}
	return results, errors.Join(failures...)
}
