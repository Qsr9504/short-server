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