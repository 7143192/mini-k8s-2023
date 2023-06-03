## 网络通信相关配置步骤
### 1. 下载并安装etcd
````
$ wget https://github.com/etcd-io/etcd/releases/download/v3.4.25/etcd-v3.4.25-linux-amd64.tar.gz
$ tar -xzvf etcd-v3.4.25-linux-amd64.tar.gz
# 将etcd和etcdctl移动至/usr/local/bin目录下。便于使用
cd etcd-v3.4.25-linux-amd64.tar.gz
mv etcd /usr/local/bin
mv etcdctl /usr/local/bin/
# 运行etcd
etcd 
````

### 2. 下载并安装flannel
````
$ wget https://github.com/flannel-io/flannel/releases/latest/download/flanneld-amd64 && chmod +x flanneld-amd64
````
在https://github.com/flannel-io/flannel/releases中可下载压缩包，其中有mk-docker-opts.sh脚本可以生成docker启动参数，默认生成位置为/run/docker_opts.env，
在docker服务启动文件中设置EnvironmentFile=/run/docker_opts.env


### 3. 在etcd中进行flannel相关配置
````
eg:
$ etcdctl put /coreos.com/network/config '{ "Network": "10.5.0.0/16", "Backend": {"Type": "vxlan"}}'
````

### 4. 启动flannel
````
$ sudo ./flanneld-amd64
````

### 5. 查看flannel相关配置
````
$ cat /run/flannel/subnet.env
FLANNEL_NETWORK=10.5.0.0/16
FLANNEL_SUBNET=10.5.86.1/24
FLANNEL_MTU=1450
FLANNEL_IPMASQ=false
$ etcdctl get /coreos.com/network/subnets --prefix --keys-only
/coreos.com/network/subnets/10.5.86.0-24
$ etcdctl get /coreos.com/network/subnets/10.5.86.0-24
/coreos.com/network/subnets/10.5.86.0-24
{"PublicIP":"192.168.174.128","PublicIPv6":null,"BackendType":"vxlan","BackendData":{"VNI":1,"VtepMAC":"5a:8a:d8:d7:c2:ed"}}
$ ./mk-docker-opts.sh -c
$ cat /run/docker_opts.env
DOCKER_OPTS=" --bip=10.5.86.1/24 --ip-masq=true --mtu=1450"
````

### 6. 修改docker配置文件进行flannel配置
首先运行下载的flannel文件夹下的`mk-docker-opts.sh`脚本，其会将当前存储在etcd中的状态转化成docker启动需要的参数形式。(所以需要在etcd开启的情况下运行)。

之后到`cat `目录下，编辑docker.service文件，添加如下内容：
```
EnvironmentFile=/run/docker_opts.env(这个路径就是上一步生成的env文件的默认路径)
ExecStart=/usr/bin/containerd $DOCKER_OPTS -H fd:// --.......(后面的不动)
```

之后保存退出，先检查/etc/docker文件夹下是否有`daemon.json`文件，如果没有则创建，并将内容设置为{}。

最后运行如下指令重启docker服务：(注：由于每次关机之后flannel会自动关闭，所以每次重启之后需要重复上述操作)

```
sudo systemctl daemon-reload
sudo systemctl restart docker
```

