package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/axgle/mahonia"
)

type SendJson struct {
	SendType    string
	CommandName string
	Params      []string
}
type Conn struct {
	C net.Conn
	WriteData chan []byte
	ReadData chan []byte
}
var conns map[string]Conn

const version = "1.0"

func main() {
	conns = make(map[string]Conn)
	// tcp 监听并接受端口
	l, err := net.Listen("tcp", "0.0.0.0:10000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	//最后关闭
	defer l.Close()
	fmt.Println("tcp服务端开始监听10000端口...")
	// 使用循环一直接受连接
	go func() {
		//从命令行接收指令
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			line := input.Text()
			// 输入exit时 结束
			if line == "exit" {
				break
			}
			arr := strings.Split(line, " ")
			for index, item := range arr {
				if item == "" {
					arr = append(arr[:index], arr[index+1:]...)
				}
			}
			fmt.Println(arr)
			if len(arr) > 0 {
				switch arr[0] {
				case "-help":
					fmt.Println("-a                       获取所有连接信息")
					fmt.Println("-t [ip:port] [Command]   向指定连接发送信息")
					fmt.Println("    if CommandName== -v      返回客户端版本信息")
					fmt.Println("    if CommandName== -exit   关闭客户端")
					fmt.Println("-v                       获取服务端版本信息")
					fmt.Println("[Command]                向全部连接发送信息")
				case "-a":
					GetAllConn()
				case "-t":
					var send SendJson
					send.SendType = "111"
					send.CommandName = arr[2]
					send.Params = arr[3:]
					SendTo(arr[1], send)
				case "-v":
					fmt.Println("server ", version)
				default:
					var send SendJson
					send.SendType = "111"
					send.CommandName = arr[0]
					send.Params = arr[1:]
					Broadcasting(send)
				}
			}
		}
		panic("exit")
	}()
	for {
		//Listener.Accept() 接受连接
		c, err := l.Accept()
		if err != nil {
			return
		}
		//处理tcp请求
		go handleConnection(c)
	}
}

//用于读取客户端传输的数据
func handleConnection(c net.Conn) {
	// fmt.Println("tcp服务端开始处理请求...")
	fmt.Println(c.RemoteAddr().String() + " 连接...")

	conn:=Conn{
		C:c,
		WriteData:make(chan []byte),
		ReadData:make(chan []byte),
	}
	conns[c.RemoteAddr().String()]=conn
	defer func(conn net.Conn) {
		delete(conns, c.RemoteAddr().String())
		conn.Close()
	}(c)
	r := bufio.NewReader(c)

	ctx,cancel:=context.WithCancel(context.Background())
	//写管道

	//读协程
	go func(contx context.Context,thisconn Conn,cancelfunc context.CancelFunc) {
		for {
			select {
			case <-contx.Done():
				goto OUTLOOP
			default:
				//读取客户端数据
				data, err := r.ReadBytes('\n')
				if err != nil || io.EOF == err {
					fmt.Println(err)
					cancelfunc()

				}
				thisconn.ReadData<-data
			}
		}
		OUTLOOP:
			fmt.Println("读协程已结束")
	}(ctx,conn,cancel)
	//ping:=[]byte(`PING\n`)
	//处理读取的数据和心跳
	go func(contx context.Context,cancelfunc context.CancelFunc,thisconn Conn) {
		timer:=time.AfterFunc(30*time.Second, func() {
			cancelfunc()
		})
		for{
			select {
			case <-contx.Done():
				goto OUTLOOP
			case data:=<-thisconn.ReadData:
				if string(data) =="PING\n" {
					timer.Stop()
					timer=time.AfterFunc(30*time.Second, func() {
						cancelfunc()
					})
					//timer.Reset(20*time.Second)
					thisconn.WriteData<-[]byte(`PONG`)
					continue
				}
				fmt.Print(ConvertToString(string(data), "GBK", "UTF-8"))
			}
		}
	OUTLOOP:
		timer.Stop()
		fmt.Println("处理读取的数据和心跳结束")
	}(ctx,cancel,conn)
	//处理写入数据
	go func(contx context.Context,thisconn Conn) {
		for{
			select {
			case <-contx.Done():
				goto OUTLOOP
			case data:=<-thisconn.WriteData:
				thisconn.C.Write(append(data,'\n'))

			}
		}
	OUTLOOP:
		fmt.Println("处理写入数据结束")
	}(ctx,conn)
	<-ctx.Done()
	conn.C.Close()
	fmt.Println(c.RemoteAddr().String() + " 断开连接...")
}

//ConvertToString 编码转换 src是待转换的字符串 srccode是src的编码 tagcode是要转成的编码
func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

//GetAllConn 获取所有连接
func GetAllConn() {
	var i int
	for key := range conns {
		i++
		fmt.Println(i, ". ", key)
	}
}

//SendTo 向指定连接发送指令
func SendTo(who string, send SendJson) {
	conn, ok := conns[who]
	if ok {
		bt, _ := json.Marshal(send)
		conn.WriteData<-append(bt, '\n')

	} else {
		fmt.Println("The " + who + " doesn't exist")
	}
}

//Broadcasting 用于向所有客户端发送指令
func Broadcasting(send SendJson) {
	var i int
	for _, conn := range conns {
		if conn.C != nil {
			bt, _ := json.Marshal(send)
			conn.WriteData<-append(bt, '\n')
			i++
		}
	}
	fmt.Println("send num:", i)
}
