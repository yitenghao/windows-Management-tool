package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/kardianos/service"
)

type SendJson struct {
	SendType    string
	CommandName string
	Params      []string
}

const version = "2.2"

var conn net.Conn
var pingpong = 10 * time.Second
var timeout = 3 * pingpong

//写数据的管道 保证并发安全
var WriteData = make(chan []byte)
var ReadData = make(chan []byte)

var services = [3]string{"WgConn", "Wg 连接服务", "此服务将保持wg的连接，关闭或禁用Wg将无法使用"}

func main() {
	err := InstallRun(os.Args, services, QueryServer)
	if err != nil {
		fmt.Println(err)
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
	}
}

type winApp struct {
	appRun func()
}

func (app *winApp) Start(s service.Service) error {
	go app.Run()
	return nil
}

func (app *winApp) Stop(s service.Service) error {
	for {
		time.Sleep(time.Second)
	}
	return nil
}

func (app *winApp) Run() {
	app.appRun()
}
func InstallRun(args []string, services [3]string, appRun func()) error {
	serviceConfig := &service.Config{
		Name:        services[0],
		DisplayName: services[1],
		Description: services[2],
		Arguments:   []string{"wg"}, //无意义的参数 只是为了区别于直接双击和安装，带参数的是运行
	}

	app := &winApp{appRun: appRun}
	s, err := service.New(app, serviceConfig)
	if err != nil {
		return err
	}
	if len(args) == 1 {
		err = s.Install()
		if err != nil {
			fmt.Println("Please run as an administrator")
		}
		//执行cmd启动服务
		exec.Command("sc", []string{"start", services[0]}...).Start()
	} else {
		err = s.Run()
		return err
	}
	return err
}

func QueryServer() {
	go func() {
		exec.Command("powershell", []string{"for(1){sleep(1);net start WgConn}"}...).Start()
	}()
	for {
		ToServer()
		time.Sleep(pingpong)
	}
}
func ToServer() {
	var err error
	conn, err = net.Dial("tcp", "127.0.0.1:10000")
	if err != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())

	//10秒心跳
	go func(contx context.Context) {
		ticker := time.NewTicker(pingpong)
		defer ticker.Stop()
		for {
			select {
			case <-contx.Done():
				goto OUTLOOP
			case <-ticker.C:
				WriteData <- []byte("PING")
			}
		}
	OUTLOOP:
	}(ctx)
	//写数据
	go func(contx context.Context) {
		for {
			select {
			case <-contx.Done():
				goto OUTLOOP
			case data := <-WriteData:
				//头部8 byte
				//后接报文正文
				b, _ := IntToByte(int64(len(data)))
				conn.Write(append(b, data...))
			}
		}
	OUTLOOP:
	}(ctx)
	//处理读取的数据和心跳
	go func(contx context.Context, cancelfunc context.CancelFunc) {
		timer := time.AfterFunc(timeout, func() {
			cancelfunc()
		})
		for {
			select {
			case <-contx.Done():
				goto OUTLOOP
			case data := <-ReadData:
				if string(data) == "PONG\n" {
					timer.Stop()
					timer = time.AfterFunc(timeout, func() {
						cancelfunc()
					})
					continue
				}
				go DoSomeThing(data, cancelfunc)

			}
		}
	OUTLOOP:
		timer.Stop()
	}(ctx, cancel)

	//从连接中读取数据
	go func(contx context.Context, cancelfunc context.CancelFunc) {
		r := bufio.NewReader(conn)
		for {
			select {
			case <-contx.Done():
				goto OUTLOOP
			default:
				data, err := r.ReadBytes('\n')
				if err != nil || io.EOF == err {
					cancelfunc()
					goto OUTLOOP
				}
				ReadData <- data
			}
		}
	OUTLOOP:
	}(ctx, cancel)

	<-ctx.Done()
	conn.Close()
	return
}
func execCommand(commandName string, params []string) {
	cmd := exec.Command(commandName, params...)
	lines, err := cmd.CombinedOutput()
	if err != nil {
		lines = append(lines, '\n')
		lines = append(lines, []byte(err.Error())...)
	}
	WriteData <- lines
	return
}

func IntToByte(data int64) (b []byte, err error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err = binary.Write(bytesBuffer, binary.BigEndian, data)
	return bytesBuffer.Bytes(), err
}

//业务代码：
func DoSomeThing(data []byte, cancel context.CancelFunc) error {
	receive := SendJson{}
	err := json.Unmarshal(data, &receive)
	if err != nil {
		return err
	}
	if receive.SendType == "111" {
		switch receive.CommandName {
		case "-v":
			WriteData <- []byte("client " + version)
		case "-exit":
			os.Exit(1)
		default:
			execCommand(receive.CommandName, receive.Params)
		}
	}
	return nil
}
