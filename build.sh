server_name=short-server

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -tags jsoniter -o mamba-short-server .
upx -9 mamba-short-server -o short-server

cat > ./Dockerfile << EOF
FROM alpine:latest
RUN apk update && apk add busybox-extras tzdata
WORKDIR /app
# 将编译好的程序拷贝进容器中
COPY ./short-server .
COPY ./config.yaml /app
ENTRYPOINT ["./short-server"]
EOF
docker build -t ${server_name} .
docker tag ${server_name} qsr9504/short-server:latest
docker tag ${server_name} qsr9504/short-server:1.0.0
docker push qsr9504/short-server:latest
docker push qsr9504/short-server:1.0.0

docker rmi qsr9504/short-server:latest
docker rmi qsr9504/short-server:1.0.0
docker rmi ${server_name}
rm -f mamba-short-server
rm -f Dockerfile
rm -f short-server