# DDNS

> 一个好用的DDNS工具

## 主要功能

​	为您的设备提供动态解析到Cloudflare

​	无需手动添加解析，本程序会自动处理，您只需输入类似于xxx.xxx.xxx的域名即可

## 安装说明

### 安装

​	您可以根据您的操作系统和处理器架构选择已经编译好了的程序

### 源代码编译

​	此方式需要您的电脑已经安装golang

```shell
# 进入项目目录
go build -o DDNS
#给DDNS可执行权限 (Linux需要)
chmod +x DDNS
```

## 使用说明

### 1.直接使用

```bash
» ./DDNS
欢迎使用DDNS Tools
您目前的ipv6:xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
选择服务:
1.Cloudflare
输入序号:1
您已选择Cloudflare
Cloudflare Email:xxx@example.com
Cloudflare API KEY:yourApiKey
域名(如果有多个逗号隔开):test@example.com,test1@example.com,test2@example.com
已将 xxx.example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
已将 xxxx.example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
```

稍等片刻，待程序提示解析成功并退出即可

### 2.保存配置使用

将以下配置放在`config.json`里

```json
{
    "cloudflare" : {
        "email" : "xxx@example.com",
        "apikey" : "Your Cloudflare Global API KEY",
        "domainList" : [
            
        ]
    }
}
```

把批量解析的域名放在`domainList`里面就行

```bash
» ./DDNS
欢迎使用DDNS Tools
您目前的ipv6:xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
选择服务:
1.Cloudflare
输入序号:1
您已选择Cloudflare
检测到已有配置,是否使用(y/n):y
已将 xxx.example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
已将 xxxx.example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
```

### 3.跳过配置一行命令执行

#### 没有配置或想一次性执行

```bash
» ./DDNS -service cloudflare -cfemail xxx@example.com -cfapikey YourCloudflareGlobalApiKey -domainList xxx@example.com,xxx1@example.com
欢迎使用DDNS Tools
您目前的ipv6:xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
当前服务:cloudflare
已将 xxx@example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
已将 xxx1@example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
```

#### 已有配置

````bash
» ./DDNS -service cloudflare
欢迎使用DDNS Tools
您目前的ipv6:xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
当前服务:cloudflare
已将 xxx@example.com 解析到 xxxx:xxx:xxxx:xxxx:xx:xxxx:xxxx:xxxx
````



### 以上就是使用方法，欢迎给个star