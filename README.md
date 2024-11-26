# 短连接服务
## 一、编辑配置信息
在当前目录下配置一个 redis 和相关启动信息

## 二、接口说明
获取短连接
```bash
curl --location 'localhost:7777/shorten' \
--header 'Content-Type: application/json' \
--data '{
    "long_url":"https://www.baidu.com/s?wd=hello"
}'
```

获取某一个短连接的点击次数
```bash
curl --location --request GET 'localhost:7777/stats/0af707bd' \
--header 'Content-Type: application/json' \
--data '{
    "long_url":"https://www.baidu.com/s?wd=hello"
}'
```

## 三、Docker 拉取命令
```bash
docker pull qsr9504/short-server:latest
```

docker 启动命令
```bash
docker run -d --name shortlink-server --restart=always -v ${你自己的配置文件路径}/config.yaml:/app/config.yaml -p 7777:7777 qsr9504/short-server:latest
```

配置文件模版文件
```yaml
base:
  website: 'http://192.168.5.8:7777'   # 域名 指向到该服务及其端口
  port: '7779'  # 服务启动端口
  length: 8     # 短链接秘钥生成长度
  cacheTime: 60 # 内存缓存时间，分钟

redis:
  addr: '192.168.5.8'   # 地址
  port: '6379'      # 端口号
  pwd:  'shatang'   # 密码
```