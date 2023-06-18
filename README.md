#运行注册中心 consul
consul agent -dev

#生成docker镜像
docker build --tag docker-go-product-service:latest .

#查看镜像
docker images

#启动容器
docker run -p 50051:50051 -p 8081:8080 e4ce6fe54b59(image号)

#生成proto文件
protoc --go_out=. --go-grpc_out=. product/product.proto