package str

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	workerBits    = 10 // 每台机器(节点)的ID位数，最大支持1024个节点
	numberBits    = 14 // 每个节点每秒最多生成16384个唯一ID
	timestampBits = 63 - workerBits - numberBits
	workerMax     = int64(1)<<workerBits - 1
	numberMax     = int64(1)<<numberBits - 1
	timestampMax  = int64(1)<<timestampBits - 1
	timeShift     = workerBits + numberBits
	workerShift   = numberBits

	// 使用秒级时间戳，保证ID始终为正数且不超过有符号int64最大值。
	// 2024-01-01 00:00:00 UTC，可用约17438年。
	epoch int64 = 1704067200
)

var (
	instance     *Worker
	instanceOnce sync.Once
)

// GetSnowWorkIns 获取静态对象
func GetSnowWorkIns() *Worker {
	instanceOnce.Do(func() {
		instance, _ = NewSnowWorker((&Worker{}).getWorkerId())
	})
	return instance
}

// Worker 定义一个worker工作节点所需要的基本参数
type Worker struct {
	mu        sync.Mutex
	timestamp int64
	workerId  int64
	number    int64
}

// NewSnowWorker 实例化一个工作节点
func NewSnowWorker(workerId int64) (*Worker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("worker id excess of quantity")
	}
	return &Worker{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

// GetId 获取唯一id
func (w *Worker) GetId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.currentSecond()
	if now < w.timestamp {
		now = w.timestamp
	}
	if w.timestamp == now {
		w.number++
		if w.number > numberMax {
			now = w.nextSecond(w.timestamp)
			w.number = 0
		}
	} else {
		w.number = 0
	}
	if now > timestampMax {
		panic("snowflake timestamp excess of quantity")
	}
	w.timestamp = now
	return now<<timeShift | (w.workerId << workerShift) | w.number
}

func (w *Worker) nextSecond(last int64) int64 {
	now := w.currentSecond()
	for now <= last {
		time.Sleep(time.Millisecond)
		now = w.currentSecond()
	}
	return now
}

func (w *Worker) currentSecond() int64 {
	now := time.Now().Unix() - epoch
	if now < 0 {
		return 0
	}
	return now
}

func (w *Worker) getWorkerId() int64 {
	values := make([]string, 0)
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		values = append(values, "hostname="+hostname)
	}
	values = append(values, "pid="+strconv.Itoa(os.Getpid()))
	if executable, err := os.Executable(); err == nil && executable != "" {
		values = append(values, "executable="+executable)
	}
	if interfaces, err := net.Interfaces(); err == nil {
		for _, item := range interfaces {
			if item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
				continue
			}
			if mac := item.HardwareAddr.String(); mac != "" {
				values = append(values, "mac="+mac)
			}
			address, err := item.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range address {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil || ip == nil || ip.IsLoopback() {
					continue
				}
				values = append(values, "ip="+ip.String())
			}
		}
	}
	if len(values) == 0 {
		values = append(values, "default")
	}
	sort.Strings(values)

	hash := fnv.New64a()
	_, _ = hash.Write([]byte(strings.Join(values, "|")))
	return int64(hash.Sum64() % uint64(workerMax+1))
}

// md5 获取md5值
func (w *Worker) md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// GetUuid 获取uuid
func (w *Worker) GetUuid() string {
	id := w.md5(strconv.FormatInt(w.GetId(), 10))
	id = fmt.Sprintf("%s-%s-%s-%s-%s", id[0:6], id[6:10], id[10:14], id[14:20], id[20:32])
	return id
}
