#集成coredns-redisc部署

##快速使用

- 1 源码下载https://github.com/coredns/coredns (因为要使用外部插件，所以需要下载源码来编译，不能使用官方提供的二进制文件)
- 2 编辑配置文件plugin.cfg，在合适的位置添加我们的插件(位置决定插件的执行顺序)，这里我们建议放在forward上面

```
......
coredns-redisc:github.com/wl4g/coredns-redisc
forward:forward
......
```

- 3 编译，在执行make之前，可以修改Makefile来修改配置，实现交叉编译，例如：
```
在SYSTEM:=后面追加"GOOS=linux GOARCH=amd64",则生成的是linux系统的二进制文件:
SYSTEM:=GOOS=linux GOARCH=amd64
SYSTEM:=GOOS=windows GOARCH=amd64
SYSTEM:=GOOS=darwin GOARCH=amd64
```

- 4 配置Corefile配置文件，具体配置方式可去coredns官网查看，这里只给个例子：
```
.:53 {
    # Load local /etc/hosts
    hosts {
        fallthrough
    }
    coredns-redisc {
        address localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381
        password "123456"
        connect_timeout 1000
        read_timeout 1000
        ttl 360
        prefix _dns:
    }
    forward . 47.107.57.204 8.8.8.8 114.114.114.114
    log
}
```

> hosts和forward: 都是coredns官方插件
> coredns-redisc: 是我们的插件名称
> address: 是redis的地址
> password: 是密码
> connect_timeout: 是redis的连接超时时间
> read_timeout: 是redis的读取超时时间
> prefix: 是存在redis的key前缀

- 启动，编译后会生成coredns可执行文件，直接执行即可，默认读取当前目录下的Corefile，也可以通过指定

```
  -conf string
    	Corefile to load (default "Corefile")
  -dns.port string
    	Default port (default "53")
  -pidfile string
    	Path to write pid file
  -plugins
    	List installed plugins
  -quiet
    	Quiet mode (no initialization output)
  -version
    	Show version
```

##二次开发

- 在coredns的plugin目录下直接拉取我们的插件
- 修改配置文件plugin.cfg: 
```
coredns-redisc:github.com/wl4g/coredns-redisc
改成: 
coredns-redisc:coredns-redisc
```

- 用开发工具直接打开coredns工程，启动直接运行coredns.go的mian即可
