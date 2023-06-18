# 基础镜像
FROM golang:1.20-alpine

# 设置工作目录
WORKDIR /app

# 复制项目文件到容器中
COPY . .

# 安装 MySQL 客户端
RUN apk update && apk add --no-cache mysql-client

# 构建应用
RUN go build -o main ./server/main.go

# 设置容器启动命令
CMD ["./main"]

# 暴露Http RestFul 8080,gRPC 50051
EXPOSE 8080
EXPOSE 50051