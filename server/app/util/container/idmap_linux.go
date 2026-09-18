//go:build linux && cgo

package container

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	userUtil "github.com/moby/sys/user"
	"golang.org/x/sys/unix"
)

// allocateIDMap 复用 root 的映射；首次使用时直接补齐系统配置，不依赖 usermod。
func (m *Manager) allocateIDMap(path string, size int64) (int64, error) {
	if size < defaultIDMapSize || size > (math.MaxInt32+1)/2 {
		return 0, errors.New("容器 UID/GID 映射数量超出系统范围")
	}
	root, err := userUtil.LookupUid(0)
	if err != nil {
		return 0, err
	}
	var data []byte
	var values []userUtil.SubID
	locked := false
	for {
		data, err = os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return 0, err
		}
		values = nil
		for index, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Split(line, ":")
			if len(fields) != 3 || fields[0] == "" {
				return 0, fmt.Errorf("%s 第 %d 行映射格式无效", path, index+1)
			}
			start, startErr := strconv.ParseUint(fields[1], 10, 32)
			count, countErr := strconv.ParseUint(fields[2], 10, 32)
			if startErr != nil || countErr != nil || count == 0 || start+count > math.MaxUint32+1 {
				return 0, fmt.Errorf("%s 第 %d 行映射范围无效", path, index+1)
			}
			values = append(values, userUtil.SubID{Name: fields[0], SubID: int64(start), Count: int64(count)})
		}
		for _, value := range values {
			if (value.Name == root.Name || value.Name == "0") && value.Count >= size &&
				value.SubID >= size && value.SubID <= math.MaxInt32-size+1 {
				return value.SubID, nil
			}
		}
		// 已有映射只读即可；缺少时先加锁，再重新读取一次。
		if locked {
			break
		}
		// 与 shadow/usermod 使用同一锁协议：先写完整 PID，再用硬链接原子占锁。
		lock, err := os.CreateTemp(filepath.Dir(path), ".subid-lock-")
		if err != nil {
			return 0, err
		}
		defer os.Remove(lock.Name())
		_, writeErr := fmt.Fprintf(lock, "%d\x00", os.Getpid())
		err = errors.Join(writeErr, lock.Sync(), lock.Close())
		if err != nil {
			return 0, err
		}
		deadline := time.Now().Add(5 * time.Second)
		for {
			if err = os.Link(lock.Name(), path+".lock"); err == nil {
				break
			}
			if !errors.Is(err, os.ErrExist) || time.Now().After(deadline) {
				return 0, fmt.Errorf("锁定 %s 失败: %w", path, err)
			}
			time.Sleep(50 * time.Millisecond)
		}
		defer os.Remove(path + ".lock")
		locked = true
	}

	// 避开所有已分配区间，以及本机已有账号或组实际使用的 ID。
	if filepath.Base(path) == "subuid" {
		users, err := userUtil.ParsePasswdFile(filepath.Join(filepath.Dir(path), "passwd"))
		if err != nil {
			return 0, err
		}
		for _, user := range users {
			values = append(values, userUtil.SubID{SubID: int64(user.Uid), Count: 1})
		}
	} else {
		groups, err := userUtil.ParseGroupFile(filepath.Join(filepath.Dir(path), "group"))
		if err != nil {
			return 0, err
		}
		for _, group := range groups {
			values = append(values, userUtil.SubID{SubID: int64(group.Gid), Count: 1})
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].SubID < values[j].SubID })
	start := max(int64(100000), size)
	for _, value := range values {
		if start+size <= value.SubID {
			break
		}
		if end := value.SubID + value.Count; start < end {
			start = end
		}
	}
	if start > math.MaxInt32-size+1 {
		return 0, errors.New("没有足够的空闲 UID/GID 用于容器映射")
	}

	// 仅追加一条记录，保留已有文本、权限、属主及 SELinux 等扩展属性。
	entry := fmt.Sprintf("%s:%d:%d\n", root.Name, start, size)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		entry = "\n" + entry
	}
	fd, err := unix.Open(path, unix.O_WRONLY|unix.O_CREAT|unix.O_APPEND|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0644)
	if err != nil {
		return 0, err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	if !info.Mode().IsRegular() || info.Size() != int64(len(data)) {
		return 0, fmt.Errorf("%s 不是普通文件或已被其他进程修改", path)
	}
	written, err := file.WriteString(entry)
	if err == nil && written != len(entry) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = file.Sync()
	}
	if err != nil {
		return 0, errors.Join(err, file.Truncate(int64(len(data))), file.Sync())
	}
	return start, nil
}
