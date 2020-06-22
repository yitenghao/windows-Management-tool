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
	"os/exec"
	"time"
)

type SendJson struct {
	SendType    string
	CommandName string
	Params      []string
}

const version = "1.1"
var conn net.Conn
var pingpong=10*time.Second
var timeout=3*pingpong
//写数据的管道 保证并发安全
var WriteData =make(chan []byte)
var ReadData =make(chan []byte)
func main() {
	for {
		ToServer()
		time.Sleep(pingpong)
	}
}
func ToServer() {
	var err error
	conn, err = net.Dial("tcp", "127.0.0.1:10000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("已连接")
	ctx,cancel:=context.WithCancel(context.Background())

	//10秒心跳
	go func(contx context.Context){
		ticker:=time.NewTicker(pingpong)
		defer ticker.Stop()
		for{
			select{
			case <-contx.Done():
				goto OUTLOOP
			case <-ticker.C:
				WriteData<-[]byte("PING")
			}
		}
	OUTLOOP:
		fmt.Println("心跳已结束")
	}(ctx)
	//写数据
	go func(contx context.Context){
		for{
			select{
			case <-contx.Done():
				goto OUTLOOP
			case data:=<-WriteData:
				//头部5 byte
				//后接报文正文
				b,_:=IntToByte(int64(len(data)))
				fmt.Printf("bytes: % x \n", b)
				conn.Write(append(b,data...))

			}
		}
		OUTLOOP:
			fmt.Println("写数据已结束")
	}(ctx)
	//处理读取的数据和心跳
	go func(contx context.Context,cancelfunc context.CancelFunc){
		timer:=time.AfterFunc(timeout, func() {
			cancelfunc()
			fmt.Println("超时")
		})
		for{
			select{
			case <-contx.Done():
				goto OUTLOOP
			case data:=<-ReadData:
				if string(data)=="PONG\n"{
					fmt.Println(string(data))
					timer.Stop()
					timer=time.AfterFunc(timeout, func() {
						cancelfunc()
						fmt.Println("超时")
					})
					continue
				}

				receive := SendJson{}
				err = json.Unmarshal(data, &receive)
				if err != nil {
					continue
				}
				if receive.SendType == "111" {
					switch receive.CommandName {
					case "-v":
						WriteData<-[]byte("client " + version)
					case "-exit":
						cancelfunc()
					default:
						execCommand(receive.CommandName, receive.Params)
					}
				}
			}
		}
		OUTLOOP:
			timer.Stop()
		fmt.Println("处理读取的数据已结束")
	}(ctx,cancel)

	//从连接中读取数据
	go func(contx context.Context,cancelfunc context.CancelFunc) {
		r := bufio.NewReader(conn)
		for{
			select{
			case <-contx.Done():
				goto OUTLOOP
			default:
				data, err := r.ReadBytes('\n')
				if err != nil || io.EOF == err {
					fmt.Println(err)
					cancelfunc()
					goto OUTLOOP
				}

				ReadData<-data
			}
		}
		OUTLOOP:
			fmt.Println("读取连接中的数据已结束")
	}(ctx,cancel)

	<-ctx.Done()
	fmt.Println("断开连接")
	conn.Close()
	return
}
func execCommand(commandName string, params []string ) {
	fmt.Println(commandName,params)
	cmd := exec.Command(commandName, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		WriteData<-[]byte(err.Error())
		return
	}
	reader := bufio.NewReader(stdout)
	lines := []byte{}
	for {
		line, err2 := reader.ReadBytes('\n')
		lines = append(lines, line...)
		if err2 != nil || io.EOF == err2 {
			break
		}
	}
	WriteData<-lines
	cmd.Wait()
	return
}

func IntToByte(data int64)(b []byte,err error){
	bytesBuffer := bytes.NewBuffer([]byte{})
	err=binary.Write(bytesBuffer, binary.BigEndian, data)
	return bytesBuffer.Bytes(),err
}