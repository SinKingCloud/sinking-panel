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
	address := server
	if host, port, err := net.SplitHostPort(address); err == nil {
		number, err := strconv.Atoi(port)
		if host == "" || err != nil || number < 1 || number > 65535 {
			return 0, errors.New("NTP 服务器端口无效")
		}
	} else {
		host := strings.Trim(server, "[]")
		if strings.Contains(host, ":") {
			if _, err = netip.ParseAddr(host); err != nil {
				return 0, errors.New("NTP 服务器地址无效")
			}
		}
		address = net.JoinHostPort(host, "123")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	connection, err := (&net.Dialer{}).DialContext(ctx, "udp", address)
	if err != nil {
		return 0, err
	}
	defer connection.Close()
	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err = connection.SetDeadline(deadline); err != nil {
		return 0, err
	}
	var request [48]byte
	request[0] = 4<<3 | 3 // NTPv4 客户端。
	// 服务端原样回显请求时间戳；使用随机值匹配响应，不依赖错误的本机日期。
	if _, err = rand.Read(request[40:48]); err != nil {
		return 0, err
	}
	sent := time.Now()
	if _, err = connection.Write(request[:]); err != nil {
		return 0, err
	}
	var packet [512]byte
	n, err := connection.Read(packet[:])
	received := time.Now()
	if err != nil {
		return 0, err
	}
	if n < 48 {
		return 0, errors.New("NTP 响应不完整")
	}
	version, mode := packet[0]>>3&7, packet[0]&7
	if (version != 3 && version != 4) || mode != 4 || !bytes.Equal(packet[24:32], request[40:48]) {
		return 0, errors.New("NTP 响应与请求不匹配")
	}
	if packet[1] == 0 {
		return 0, fmt.Errorf("NTP 服务器拒绝请求: %q", string(packet[12:16]))
	}
	if packet[0]>>6 == 3 || packet[1] > 15 {
		return 0, errors.New("NTP 服务器尚未同步")
	}
	if binary.BigEndian.Uint64(packet[32:40]) == 0 || binary.BigEndian.Uint64(packet[40:48]) == 0 {
		return 0, errors.New("NTP 服务器返回了空时间戳")
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
		return 0, errors.New("测量期间系统时间已变化，请重试")
	}
	processing := serverSent.Sub(serverReceived)
	delay := roundTrip - processing
	if processing < 0 || delay < -time.Millisecond {
		return 0, errors.New("NTP 时间戳顺序或网络延迟异常")
	}
	if delay < 0 {
		delay = 0
	}
	rootDelay := time.Duration(int32(binary.BigEndian.Uint32(packet[4:8]))) * time.Second / 65536
	dispersion := time.Duration(binary.BigEndian.Uint32(packet[8:12])) * time.Second / 65536
	if dispersion+(max(rootDelay, 0)+delay)/2 > 5*time.Second {
		return 0, errors.New("NTP 时间源误差过大")
	}
	return serverSent.Add(delay / 2).Sub(received), nil
}
