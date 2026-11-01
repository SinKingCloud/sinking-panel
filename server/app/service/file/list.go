package file

import (
	"errors"
	"path/filepath"
	"runtime"
	"server/app/util/file"
	"server/app/util/page"
	"slices"
	"strings"
)

// List 分页浏览目录，Windows 的 / 作为磁盘列表入口。
func (s *service) List(path, keyword string, query *page.Query) (*page.Result[*File], error) {
	if query == nil || query.Page < 1 || query.PageSize < 1 {
		return nil, errors.New("无效的分页参数")
	}
	if runtime.GOOS != "windows" || filepath.Clean(path) != string(filepath.Separator) {
		items, total, err := file.NewDisk("").FileListWithPage(path, query.Page, query.PageSize, query.OrderByField, query.OrderByType, keyword)
		if err != nil {
			return nil, err
		}
		list := make([]*File, 0, len(items))
		for _, item := range items {
			list = append(list, &File{File: *item})
		}
		return page.New(query, total, list), nil
	}

	paths, err := s.GetDiskPaths()
	if err != nil {
		return nil, err
	}
	list := make([]*File, 0, len(paths))
	keyword = strings.ToLower(keyword)
	for _, path := range paths {
		name := strings.TrimSuffix(path, "/")
		name = name[strings.LastIndexByte(name, '/')+1:]
		if keyword != "" && !strings.Contains(strings.ToLower(path), keyword) {
			continue
		}
		list = append(list, &File{File: file.File{Name: name, IsDir: true}, Path: path})
	}
	if query.OrderByField == "name" {
		slices.SortFunc(list, func(a, b *File) int {
			order := strings.Compare(a.Name, b.Name)
			if query.OrderByType == "desc" {
				return -order
			}
			return order
		})
	}
	total := len(list)
	if total == 0 || query.Page-1 > (total-1)/query.PageSize {
		return page.New(query, int64(total), []*File{}), nil
	}
	start := (query.Page - 1) * query.PageSize
	end := start + min(query.PageSize, total-start)
	return page.New(query, int64(total), list[start:end]), nil
}
