package command

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"

	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/validator"
	"server/bootstrap"
)

// updateAccount 共用账号和密码修改流程，未修改的字段保持原值。
func (s *Server) updateAccount(command string) error {
	if err := s.requireInteractive(command); err != nil {
		return err
	}
	var account string
	var password []byte
	var err error
	label, missingLabel, missingKey, missingCommand := "账号", "密码", constant.LoginPassword, "pwd"
	if command == "user" {
		account, err = s.readAccount("请输入新账号: ")
	} else {
		label, missingLabel, missingKey, missingCommand = "密码", "账号", constant.LoginAccount, "user"
		password, err = s.readNewPassword("请输入新密码: ", "请再次输入新密码: ", false)
		defer clear(password)
	}
	if err != nil {
		return err
	}
	bootstrap.Load()
	defer bootstrap.Close()
	service.Init()
	if err = service.Auth.UpdateAccount(account, string(password)); err != nil {
		return err
	}
	service.Log.Create("127.0.0.1", log_type.EventUpdate, "修改登录信息", "通过命令行修改登录"+label)
	log.Printf("登录%s修改成功；服务运行中时，请重新启动服务使修改立即生效", label)
	if service.Config.Get(constant.LoginGroup, missingKey) == "" {
		log.Printf("登录%s尚未设置，请继续执行 %s %s\n", missingLabel, os.Args[0], missingCommand)
	}
	return nil
}

func (s *Server) requireInteractive(command string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("请在交互式终端中执行 %s 命令", command)
	}
	return nil
}

func (s *Server) readAccount(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	account, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("读取账号失败: %w", err)
	}
	if errors.Is(err, io.EOF) && account == "" {
		return "", errors.New("已取消输入")
	}
	form := struct {
		Account string `validate:"required,alphanum,max=20" label:"登录账号"`
	}{Account: strings.TrimSpace(account)}
	if ok, message := validator.Check(&form); !ok {
		return "", errors.New(message)
	}
	return form.Account, nil
}

func (s *Server) readNewPassword(prompt string, confirmPrompt string, visible bool) ([]byte, error) {
	var input *bufio.Reader
	if visible {
		input = bufio.NewReader(os.Stdin)
	}
	read := func(prompt string) ([]byte, error) {
		fmt.Fprint(os.Stderr, prompt)
		var password []byte
		var err error
		if input != nil {
			password, err = input.ReadBytes('\n')
			if errors.Is(err, io.EOF) && len(password) > 0 {
				err = nil
			}
			// 只去掉行结束符，保留密码本身的首尾空格。
			password = bytes.TrimSuffix(bytes.TrimSuffix(password, []byte{'\n'}), []byte{'\r'})
		} else {
			password, err = term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
		}
		if err != nil {
			clear(password)
			return nil, fmt.Errorf("读取密码失败: %w", err)
		}
		return password, nil
	}
	password, err := read(prompt)
	if err != nil {
		return nil, err
	}
	confirm, err := read(confirmPrompt)
	if err != nil {
		clear(password)
		return nil, err
	}
	defer clear(confirm)
	if !bytes.Equal(password, confirm) {
		clear(password)
		return nil, errors.New("两次输入的密码不一致")
	}
	if !utf8.Valid(password) {
		clear(password)
		return nil, errors.New("登录密码格式不合法")
	}
	length := utf8.RuneCount(password)
	if length < 6 || length > 20 {
		clear(password)
		return nil, errors.New("登录密码长度必须在6到20个字符之间")
	}
	if len(password) > 72 {
		clear(password)
		return nil, errors.New("登录密码不能超过72个字节")
	}
	return password, nil
}
