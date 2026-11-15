package timezone

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// query 仅测量偏差，不修改时钟。往返耗时使用单调时钟，避免本机时间不准影响测量。
func query(ctx context.Context, server string) (time.Duration, error) {
	if server == "" || strings.ContainsAny(server, " /\\\t\r\n\x00") {
		return 0, errors.New("NTP 服务器地址无效")
	}
	host, port, err := net.SplitHostPort(server)
	if err == nil {
		number, err := strconv.Atoi(port)
		if host == "" || err != nil || number < 1 || number > 65535 {
			return 0, errors.New("NTP 服务器端口无效")
		}
	} else {
		host, port = strings.Trim(server, "[]"), "123"
		if strings.Contains(host, ":") {
			if _, err = netip.ParseAddr(host); err != nil {
				return 0, errors.New("NTP 服务器地址无效")
			}
		}
	}
	// UDP connect 成功不代表时间源可达，必须展开 pool 域名的地址逐个尝试。
	lookup, cancel := context.WithTimeout(ctx, 3*time.Second)
	addresses, err := net.DefaultResolver.LookupIPAddr(lookup, host)
	lookupErr := lookup.Err()
	cancel()
	if err != nil {
		if lookupErr != nil {
			err = lookupErr
		}
		return 0, fmt.Errorf("解析 NTP 时间源失败: %w", err)
	}
	if len(addresses) == 0 {
		return 0, errors.New("NTP 时间源没有可用地址")
	}
	var families [2][]string
	seen := make(map[string]bool, len(addresses))
	for _, address := range addresses {
		endpoint := net.JoinHostPort(address.String(), port)
		if !seen[endpoint] {
			family := 0
			if (address.IP.To4() != nil) != (addresses[0].IP.To4() != nil) {
				family = 1
			}
			families[family] = append(families[family], endpoint)
			seen[endpoint] = true
		}
	}
	// 交替尝试两种地址族，避免多个不可达 IPv6 耗尽机会后还没试过 IPv4。
	endpoints := make([]string, 0, len(addresses))
	for index := 0; index < max(len(families[0]), len(families[1])); index++ {
		for _, family := range families {
			if index < len(family) {
				endpoints = append(endpoints, family[index])
			}
		}
	}
	// 每次使用新连接和随机时间戳，迟到的旧包不会被当成本次响应。
	request := func(address string) (offset time.Duration, retry bool, err error) {
		attempt, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		defer func() {
			if err == nil {
				return
			}
			if cause := attempt.Err(); cause != nil {
				err = cause
			} else if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				err = context.DeadlineExceeded
			}
			if errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("NTP 请求超时: %w", err)
			} else if errors.Is(err, context.Canceled) {
				err = fmt.Errorf("NTP 请求已取消: %w", err)
			}
		}()
		connection, err := (&net.Dialer{}).DialContext(attempt, "udp", address)
		if err != nil {
			return 0, true, err
		}
		defer connection.Close()
		stop := context.AfterFunc(attempt, func() { _ = connection.Close() })
		defer stop()
		deadline, _ := attempt.Deadline()
		if err = connection.SetDeadline(deadline); err != nil {
			return 0, false, err
		}
		var request [48]byte
		request[0] = 4<<3 | 3 // NTPv4 客户端。
		// 服务端原样回显请求时间戳；使用随机值匹配响应，不依赖错误的本机日期。
		if _, err = rand.Read(request[40:48]); err != nil {
			return 0, false, err
		}
		sent := time.Now()
		if _, err = connection.Write(request[:]); err != nil {
			return 0, true, err
		}
		var packet [512]byte
		n, err := connection.Read(packet[:])
		received := time.Now()
		if err != nil {
			return 0, true, err
		}
		if n < 48 {
			return 0, false, errors.New("NTP 响应不完整")
		}
		version, mode := packet[0]>>3&7, packet[0]&7
		if (version != 3 && version != 4) || mode != 4 || !bytes.Equal(packet[24:32], request[40:48]) {
			return 0, false, errors.New("NTP 响应与请求不匹配")
		}
		if packet[1] == 0 {
			return 0, false, fmt.Errorf("NTP 服务器拒绝请求: %q", string(packet[12:16]))
		}
		if packet[0]>>6 == 3 || packet[1] > 15 {
			return 0, false, errors.New("NTP 服务器尚未同步")
		}
		if binary.BigEndian.Uint64(packet[32:40]) == 0 || binary.BigEndian.Uint64(packet[40:48]) == 0 {
			return 0, false, errors.New("NTP 服务器返回了空时间戳")
		}
		// 根据本机附近的 NTP era 展开 32 位秒数，支持 2036 年回绕。
		decode := func(data []byte) time.Time {
			seconds := binary.BigEndian.Uint32(data[:4])
			fraction := binary.BigEndian.Uint32(data[4:])
			unix := received.Unix() + int64(int32(seconds-uint32(received.Unix()+2208988800)))
			return time.Unix(unix, int64(uint64(fraction)*1e9>>32))
		}
		serverReceived, serverSent := decode(packet[32:40]), decode(packet[40:48])
		roundTrip := received.Sub(sent)
		clockChange := received.Round(0).Sub(sent.Round(0)) - roundTrip
		if clockChange > 100*time.Millisecond || clockChange < -100*time.Millisecond {
			return 0, false, errors.New("测量期间系统时间已变化，请重试")
		}
		processing := serverSent.Sub(serverReceived)
		delay := roundTrip - processing
		if processing < 0 || delay < -time.Millisecond {
			return 0, false, errors.New("NTP 时间戳顺序或网络延迟异常")
		}
		if delay < 0 {
			delay = 0
		}
		rootDelay := time.Duration(int32(binary.BigEndian.Uint32(packet[4:8]))) * time.Second / 65536
		dispersion := time.Duration(binary.BigEndian.Uint32(packet[8:12])) * time.Second / 65536
		if dispersion+(max(rootDelay, 0)+delay)/2 > 5*time.Second {
			return 0, false, errors.New("NTP 时间源误差过大")
		}
		return serverSent.Add(delay / 2).Sub(received), false, nil
	}
	var failures []error
	blocked := make(map[string]bool)
	// 优先切换 IP；只有网络故障才重试同一 IP，拒绝或无效响应不重复请求。
	for index, attempts := 0, 0; attempts < 3 && len(blocked) < len(endpoints); index++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		endpoint := endpoints[index%len(endpoints)]
		if blocked[endpoint] {
			continue
		}
		attempts++
		offset, retry, err := request(endpoint)
		if err == nil {
			return offset, nil
		}
		failures = append(failures, fmt.Errorf("%s: %w", endpoint, err))
		if !retry {
			blocked[endpoint] = true
		}
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("NTP 查询失败（已尝试 %d 次）: %w", len(failures), errors.Join(failures...))
}
