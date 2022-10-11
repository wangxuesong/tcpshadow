# tcpshadow

内部使用的客户端协议抓包、分析工具。支持 GBase 和 PG。

# Usage

```
Usage:
  tcpshadow [command]

Available Commands:
  capture     Capture tcp data
  client      Client request
  convert     Convert between capture files and other file formats
  help        Help about any command
  server      Server response
  show        show capture file

Flags:
  -h, --help            help for tcpshadow

Use "tcpshadow [command] --help" for more information about a command.
```

## 命令

### capture

捕获通讯包。capture 命令启动一个 Proxy 服务，客户端连接 Proxy 服务，Proxy 服务连接服务器。

```
Capture tcp data

Usage:
  tcpshadow capture [flags]

Flags:
  -h, --help            help for capture
  -l, --listen string   监听地址，用于客户端连接
  -o, --output string   保存的文件名
  -p, --printPackage    抓包的过程中打印包内容
  -s, --server string   服务器地址
  -t, --type string     客户端协议类型 (缺省 "sqli"，支持 sqli、pg)
```

### show

显示抓包文件中的内容。

```
show capture file

Usage:
  tcpshadow show [flags]

Flags:
  -c, --color          显示颜色（绿色：客户端发送；黄色：服务器发送）
  -h, --help           help for show
  -i, --input string   抓包文件
  -p, --parse          解析包格式
  -r, --raw            显示原始数据
  -t, --type string    客户端协议类型 (缺省 "sqli"，支持 sqli、p)
```