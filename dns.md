## dns配置步骤（v2）
### 1. 配置coreDNS作为DNS server
````
# v1版本，将coreDNS部署在一个单独的容器中以使用53端口接收dns解析请求，
# 后续可考虑将coreDNS部署在pod上，以service形式提供dns解析服务
# 使用ubuntu镜像作为容器镜像
$ docker run --name coredns_server -it ubuntu /bin/bash
# 以下步骤在容器中进行
# 首先安装必要工具
$ apt-get update
$ apt-get install net-tools
$ apt-get install vim
$ apt-get install curl
# ifconfig查询容器ip，以作为其他pod的dns server配置ip
$ curl -O -L  https://github.com/coredns/coredns/releases/download/v1.10.1/coredns_1.10.1_linux_amd64.tgz
$ tar -zxvf coredns_1.10.1_linux_amd64.tgz
$ vim Corefile
# 初步配置如下
.:53 {
    errors
    health
    hosts {
      192.168.174.128 example.com #此处即将host域名解析为对应的nginx转发服务器ip
      fallthrough
    }
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
}
# 运行coredns
$ ./coredns --conf Corefile_path
````
目前coredns_server容器只需要一个，运行在master节点上即可，使用前需要在容器中运行./coredns,
另外，coredns_server容器的/home/volume目录挂载在master主机的/home/root/volume目录下，
Corefile配置文件要放在主机该目录下，以便在主机中直接修改配置文件


### 2. 配置nginx作为转发服务器
````
# dns server负责将host域名转化为nginx转发服务器的ip，nginx再根据url路径
# 将请求转发到对应的serviceIp:port上，因此每当有一个DNS对象被创建，就要启动
# 一台nginx转发服务器负责该DNS对象host域名下的具体路径转发
# nginx也在容器中启动，以使用80端口接收请求
$ sudo apt-get install nginx
$ sudo vim /etc/nginx/nginx.conf
# 配置例子:将example.com:80/path1的请求转发到10.5.86.8:11000
# 可添加location = /path2 配置更多子路径
http {
    server {
          listen 80;
          server_name example.com;
          location = /path1 {
              proxy_pass http://10.5.86.8:11000;
          }
        }
}
````
此处应该可以直接使用nginx镜像
目前此步骤已经改为使用go编程实现，使用的容器镜像为nginx镜像，注意在各个节点主机上
创建/home/root/nginx目录，用于放置nginx代理服务器的配置文件，配置文件在主机上创建，
然后挂载到nginx容器的/etc/nginx目录下作为nginx代理服务器的运行配置

### 3. 修改/etc/resolv.conf
````
将dns解析服务器地址配置为coreDNS的ip
$ vim /etc/resolv.conf #修改nameserver
````
此步骤已经通过go编程实现，容器中的dns解析服务器地址在创建时自动修改（目前是在pause容器
创建时修改配置，考虑改为创建pause容器后再使用echo修改）各个主机上的/etc/resolv.conf目前
需要手动修改

### 4. 验证
````
# 在后端pod根目录下创建/path1目录，放置一个简单的html文件，然后运行一个简单的http服务器
$ python3 -m http.server 11000
# curl模拟请求
$ echo -e "GET /path1 HTTP/1.1\nHost: example.com\n\n" | nc example.com 80
````

### 5. 下一步工作
将以上命令行工作转为用go编程实现