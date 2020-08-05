"# windows-Management-tool" 


go get github.com/axgle/mahonia


go get github.com/akavel/rsrc


go install github.com/akavel/rsrc


cd server


go run main.go


cd client


go run main.go



操作方法：

客户端Dial的ip修改为服务端IP


服务端运行之后，直接在控制台输入指令，如：where powershell


实际上和在本机上cmd输入指令一样的，但是有的指令（不在path下的）需要用绝对路径，如果找不到路径就执行不了。如 start www.baidu.com 指令,就打不开百度。


但是windows自带powershell，可以直接调powershell，而且powershell.exe可以接受命令行参数，且他可以执行所有cmd指令，所以可以绕过cmd，如 powershell start www.baidu.com


作为服务启动，客户端会拥有管理员权限，服务端的指令可以享受管理员身份执行


代码很简单，逻辑也不复杂，可以自己封装改动一哈。
自己看吧：）

--共同学习，共同进步--


client文件夹下的1.ico是图标文件，ico.manifest文件是清单文件，里面配置了开启图标和管理员算法提示，使用`rsrc.exe -manifest ico.manifest -o myapp.syso -ico myapp.ico`生成.syso文件，放到main所在文件夹下，build时会加载并应用，这样就不用右键以管理员运行了。
