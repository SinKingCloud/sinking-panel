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

func (s *Server) user() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("请在交互式终端中执行 user 命令")
	}
	fmt.Fprint(os.Stderr, "请输入新账号: ")
	account, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("读取账号失败: %w", err)
	}
	if errors.Is(err, io.EOF) && account == "" {
		return errors.New("已取消输入")
	}
	form := struct {
		Account string `validate:"required,alphanum,max=20" label:"登录账号"`
	}{Account: strings.TrimSpace(account)}
	if ok, message := validator.Check(&form); !ok {
		return errors.New(message)
	}
	bootstrap.Load()
	service.Init()
	if err = service.Auth.UpdateAccount(form.Account, ""); err != nil {
		return err
	}
	if err = service.Log.Create("", log_type.EventUpdate, "修改登录信息", "通过命令行修改登录账号"); err != nil {
		return fmt.Errorf("登录账号已修改，但记录操作日志失败: %w", err)
	}
	log.Println("登录账号修改成功；服务运行中时，请重新启动服务使修改立即生效")
	if service.Config.Get(constant.LoginGroup, constant.LoginPassword) == "" {
		log.Printf("登录密码尚未设置，请继续执行 %s pwd\n", os.Args[0])
	}
	return nil
}

func (s *Server) pwd() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("请在交互式终端中执行 pwd 命令")
	}
	password, err := s.readPassword("请输入新密码: ")
	if err != nil {
		return err
	}
	defer clear(password)
	confirm, err := s.readPassword("请再次输入新密码: ")
	if err != nil {
		return err
	}
	defer clear(confirm)
	if !bytes.Equal(password, confirm) {
		return errors.New("两次输入的密码不一致")
	}
	if !utf8.Valid(password) {
		return errors.New("登录密码格式不合法")
	}
	length := utf8.RuneCount(password)
	if length < 6 || length > 20 {
		return errors.New("登录密码长度必须在6到20个字符之间")
	}
	if len(password) > 72 {
		return errors.New("登录密码不能超过72个字节")
	}
	bootstrap.Load()
	service.Init()
	if err = service.Auth.UpdateAccount("", string(password)); err != nil {
		return err
	}
	if err = service.Log.Create("", log_type.EventUpdate, "修改登录信息", "通过命令行修改登录密码"); err != nil {
		return fmt.Errorf("登录密码已修改，但记录操作日志失败: %w", err)
	}
	log.Println("登录密码修改成功；服务运行中时，请重新启动服务使修改立即生效")
	if service.Config.Get(constant.LoginGroup, constant.LoginAccount) == "" {
		log.Printf("登录账号尚未设置，请继续执行 %s user\n", os.Args[0])
	}
	return nil
}

func (s *Server) readPassword(prompt string) ([]byte, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		clear(password)
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}
	return password, nil
}
