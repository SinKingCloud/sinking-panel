package oss

import (
	"context"
	"errors"
	utilOss "server/app/util/oss"
	"server/app/util/page"
	"strings"
)

// Files 使用云厂商游标分页，目录和文件合并返回，不扫描整个桶计算总数。
func (s *service) Files(ctx context.Context, id int64, prefix, delimiter string, query *page.Query) (*page.Result[*File], error) {
	if query == nil {
		query = &page.Query{PageSize: 20}
	}
	if query.Page > 0 {
		return nil, errors.New("云端文件列表仅支持游标分页")
	}
	if query.PageSize < 1 || query.PageSize > 1000 {
		return nil, errors.New("分页容量必须为 1–1000")
	}
	client, err := s.client(id)
	if err != nil {
		return nil, err
	}
	options := utilOss.ListOptions{
		Prefix:    prefix,
		Delimiter: delimiter,
		Token:     query.CursorId,
		Limit:     query.PageSize,
	}
	// 从目录名之后开始列举，由云端跳过当前目录占位对象，避免占用分页名额。
	if options.Token == "" && strings.HasSuffix(prefix, "/") {
		options.StartAfter = prefix
	}
	data, err := client.ListFiles(ctx, options)
	if err != nil {
		return nil, err
	}
	list := make([]*File, 0, len(data.CommonPrefixes)+len(data.Files))
	seen := make(map[string]struct{}, cap(list))
	for _, key := range data.CommonPrefixes {
		if _, exists := seen[key]; exists || key == prefix {
			continue
		}
		seen[key] = struct{}{}
		list = append(list, &File{
			File:      utilOss.File{Key: key},
			Directory: true,
		})
	}
	for _, item := range data.Files {
		if _, exists := seen[item.Key]; exists || (item.Key == prefix && strings.HasSuffix(prefix, "/")) {
			continue
		}
		seen[item.Key] = struct{}{}
		list = append(list, &File{
			File:      item,
			Directory: strings.HasSuffix(item.Key, "/"),
		})
	}
	query.HasMore = data.HasMore
	query.NextCursorId = data.NextToken
	query.NextCursorLastId = ""
	return page.New(query, 0, list), nil
}
