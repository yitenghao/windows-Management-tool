"# windows-Management-tool-" 
go get github.com/axgle/mahonia
cd server
go run main.go
cd client
go run main.go



操作方法：
服务端运行之后，直接在控制台输入指令，如：where powershell


实际上和在本机上cmd输入指令一样的，但是有的指令（不在path下的）需要用绝对路径，如果找不到路径就执行不了。如 start www.baidu.com 指令,就打不开百度。


但是windows自带powershell，可以直接调powershell，而且powershell.exe可以接受命令行参数，且他可以执行所有cmd指令，所以可以绕过cmd，如 powershell start www.baidu.com



代码很简单，逻辑也不复杂，可以自己封装改动一哈。
自己看吧：）

--共同学习，共同进步--
