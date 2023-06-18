package main

import (
	pb "awesomeProject/product/proto" // 替换为实际的包路径和文件名
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // 导入MySQL驱动程序包
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

// 商品结构体
type Product struct {
	ID       int64 `gorm:"primary_key"`
	Name     string
	Num      int
	Style    string
	Provider string
}

// ProductServiceServer 是商品服务的服务器
type ProductServiceServer struct {
	pb.UnimplementedProductServiceServer
}

type RequestBody struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

var db *sql.DB

func initializeDB() {
	// 从配置文件或其他来源获取MySQL连接配置
	dbHost := "host.docker.internal"
	dbPort := "3306"
	dbUser := "root"
	dbPass := "123456"
	dbName := "productservice"

	// 构建MySQL连接字符串
	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	// 建立与MySQL的连接
	var err error
	db, err = sql.Open("mysql", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}

// GetProductQuantityById 实现通过商品ID查询商品数量的RPC方法
func (s *ProductServiceServer) GetProductQuantityById(ctx context.Context, req *pb.GetProductQuantityByIdRequest) (*pb.GetProductQuantityResponse, error) {
	// 根据商品ID查询商品数量的业务逻辑
	productID := req.GetId()
	log.Println("productID", productID)
	// 在数据库中查询商品数量
	quantity := "SELECT num FROM producttable WHERE id = ?"
	//
	stmt, err := db.Prepare(quantity)
	if err != nil {
		log.Fatalf("Failed to prepare query: %v", err)
	}
	defer stmt.Close()
	//
	//// 执行查询
	var num int32
	err = stmt.QueryRow(productID).Scan(&num)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	log.Println("GetProductQuantityById获取到的数据:", num)
	return &pb.GetProductQuantityResponse{
		Quantity: num,
	}, nil

}

// GetProductQuantityByName 实现通过商品名称查询商品数量的RPC方法
func (s *ProductServiceServer) GetProductQuantityByName(ctx context.Context, req *pb.GetProductQuantityByNameRequest) (*pb.GetProductQuantityResponse, error) {
	// 根据商品名称查询商品数量的业务逻辑
	productName := req.GetName()

	// 假设在数据库中查询商品数量
	quantity := "SELECT num FROM producttable WHERE name = ?"
	stmt, err := db.Prepare(quantity)
	if err != nil {
		log.Fatalf("Failed to prepare query: %v", err)
	}
	defer stmt.Close()

	//// 执行查询
	var num int32
	err = stmt.QueryRow(productName).Scan(&num)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	log.Println("GetProductQuantityById获取到的数据:", num)
	return &pb.GetProductQuantityResponse{
		Quantity: num,
	}, nil
}

// AddProductQuantity 实现商品入库的RPC方法
func (s *ProductServiceServer) AddProductQuantity(ctx context.Context, req *pb.AddProductQuantityRequest) (*pb.AddProductQuantityResponse, error) {
	// 商品入库的业务逻辑
	productID := req.GetId()
	quantity := req.GetQuantity()

	// 更新商品数量的 SQL 语句
	updateQuery := "UPDATE producttable SET num = num + ? WHERE id = ?"
	stmt, err := db.Prepare(updateQuery)
	if err != nil {
		log.Fatalf("Failed to prepare query: %v", err)
	}
	defer stmt.Close()

	// 执行 SQL 语句，更新商品数量
	_, err = stmt.Exec(quantity, productID)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	// 重新查询数据库，获取更新后的商品数量
	selectQuery := "SELECT num FROM producttable WHERE id = ?"
	row := db.QueryRow(selectQuery, productID)
	var updatedQuantity int32
	err = row.Scan(&updatedQuantity)
	if err != nil {
		log.Fatalf("Failed to retrieve updated quantity: %v", err)
	}

	// 将更新后的数量赋值给 quantity
	quantity = updatedQuantity

	// 创建并返回响应
	response := &pb.AddProductQuantityResponse{
		StatusCode: 200,
		Message:    "商品入库成功",
		Id:         productID,
		Quantity:   quantity,
	}
	return response, nil

}

// RemoveProductQuantity 实现商品出库的RPC方法
func (s *ProductServiceServer) RemoveProductQuantity(ctx context.Context, req *pb.RemoveProductQuantityRequest) (*pb.RemoveProductQuantityResponse, error) {
	// 商品出库的业务逻辑
	productID := req.GetId()
	quantity := req.GetQuantity()

	selectQuery := "SELECT num FROM producttable WHERE id = ?"
	row := db.QueryRow(selectQuery, productID)
	var updatedQuantity int32
	err := row.Scan(&updatedQuantity)
	if err != nil {
		log.Fatalf("Failed to retrieve updated quantity: %v", err)
	}

	// 根据更新后的商品数量设置响应对象
	var response *pb.RemoveProductQuantityResponse
	var updatedNum = updatedQuantity - quantity
	if updatedNum >= 0 {
		// 更新商品数量的 SQL 语句
		updateQuery := "UPDATE producttable SET num = num - ? WHERE id = ?"
		stmt, err := db.Prepare(updateQuery)
		if err != nil {
			log.Fatalf("Failed to prepare query: %v", err)
		}
		defer stmt.Close()

		// 执行 SQL 语句，更新商品数量
		_, err = stmt.Exec(quantity, productID)
		if err != nil {
			log.Fatalf("Failed to execute query: %v", err)
		}
		response = &pb.RemoveProductQuantityResponse{
			StatusCode: 200,
			Message:    "商品出库成功",
			Id:         productID,
			Quantity:   updatedNum,
		}
	} else {
		response = &pb.RemoveProductQuantityResponse{
			StatusCode: 400,
			Message:    "商品出库失败，库存不足",
			Id:         productID,
			Quantity:   updatedQuantity,
		}

	}

	return response, nil
}

func GetProductQuantityByIdHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	productID := requestBody.ID
	log.Println("id:", productID)

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	req := &pb.GetProductQuantityByIdRequest{
		Id: productID,
	}

	resp, err := client.GetProductQuantityById(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to retrieve product quantity", http.StatusInternalServerError)
		return
	}

	response := struct {
		Quantity int32 `json:"quantity"`
	}{
		Quantity: resp.Quantity,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func GetProductQuantityByNameHandler(w http.ResponseWriter, r *http.Request) {

	var requestBody RequestBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	productName := requestBody.Name
	log.Println("id:", productName)

	// 创建 gRPC 连接
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 创建商品服务的 gRPC 客户端
	client := pb.NewProductServiceClient(conn)

	// 构建 gRPC 请求
	req := &pb.GetProductQuantityByNameRequest{
		Name: productName,
	}

	// 调用 gRPC 方法
	resp, err := client.GetProductQuantityByName(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to retrieve product quantity", http.StatusInternalServerError)
		return
	}

	// 创建响应对象
	response := struct {
		Quantity int32 `json:"quantity"`
	}{
		Quantity: resp.Quantity,
	}

	// 将响应对象转换为 JSON 格式并返回
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func AddProductQuantityHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Call WalletAmount  METHOD:%s\n", r.Method)
	log.Println("hello !!!!!!!!!!!!!!!")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch r.Method {
	case http.MethodPost:
		var requestBody RequestBody

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&requestBody)
		if err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		id := requestBody.ID
		log.Println("id:", id)
		quantityStr := requestBody.Quantity
		log.Println("quantity:", quantityStr)

		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := pb.NewProductServiceClient(conn)

		req := &pb.AddProductQuantityRequest{
			Id:       id,
			Quantity: quantityStr,
		}

		resp, err := client.AddProductQuantity(context.Background(), req)
		if err != nil {
			http.Error(w, "Failed to add product quantity", http.StatusInternalServerError)
			return
		}

		response := struct {
			StatusCode int64  `json:"status_code"`
			Message    string `json:"message"`
			Id         int64  `json:"id"`
			Quantity   int32  `json:"quantity"`
		}{
			StatusCode: resp.StatusCode,
			Message:    resp.Message,
			Id:         resp.Id,
			Quantity:   resp.Quantity,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}

}
func RemoveProductQuantityHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	id := requestBody.ID
	log.Println("id:", id)
	quantityStr := requestBody.Quantity
	log.Println("quantity:", quantityStr)

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	req := &pb.RemoveProductQuantityRequest{
		Id:       id,
		Quantity: quantityStr,
	}

	resp, err := client.RemoveProductQuantity(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to remove product quantity", http.StatusInternalServerError)
		return
	}

	response := struct {
		StatusCode int64  `json:"status_code"`
		Message    string `json:"message"`
		Id         int64  `json:"id"`
		Quantity   int32  `json:"quantity"`
	}{
		StatusCode: resp.StatusCode,
		Message:    resp.Message,
		Id:         resp.Id,
		Quantity:   resp.Quantity,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// GetAllProducts 获取所有商品数据的处理程序
func GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	// 查询所有商品数据的 SQL 语句
	query := "SELECT * FROM producttable"

	// 执行查询
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 创建商品列表
	var products []Product

	// 遍历查询结果并将数据添加到商品列表中
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Name, &product.Num, &product.Style, &product.Provider)
		if err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	// 检查是否发生了迭代错误
	err = rows.Err()
	if err != nil {
		http.Error(w, "Failed to iterate over rows", http.StatusInternalServerError)
		return
	}

	// 将商品列表转换为 JSON 格式
	jsonResponse, err := json.Marshal(products)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func handleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	initializeDB()
	defer closeDB()

	// 服务注册
	// 1.初始化 consul 配置
	consulConfig := api.DefaultConfig()
	// 2.创建 consul 对象
	consulClient, _ := api.NewClient(consulConfig)
	// 3.注册的服务配置
	reg := api.AgentServiceRegistration{
		ID:      "product",
		Name:    "productService",
		Tags:    []string{"consul", "grpc"},
		Address: "host.docker.internal",
		Port:    8080,
		Check: &api.AgentServiceCheck{
			CheckID:  "consul grpc test",
			TCP:      "host.docker.internal:8080",
			Timeout:  "1s",
			Interval: "5s",
		},
	}
	// 4. 注册 grpc 服务到 consul 上
	consulClient.Agent().ServiceRegister(&reg)

	grpcServer := grpc.NewServer()
	productService := &ProductServiceServer{}
	pb.RegisterProductServiceServer(grpcServer, productService)

	go func() {
		port := ":50051"
		listener, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Printf("gRPC server listening on port %s...\n", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	router := mux.NewRouter()
	// 创建处理跨域的 CORS 实例
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})
	// 应用 CORS 中间件到路由器
	handler := c.Handler(router)
	// 应用跨域中间件
	router.Use(handleCORS)
	router.HandleFunc("/product/num", GetProductQuantityByIdHandler).Methods("POST")
	router.HandleFunc("/products/quantity/name", GetProductQuantityByNameHandler).Methods("POST")
	router.HandleFunc("/product/add", AddProductQuantityHandler).Methods("POST")
	router.HandleFunc("/product/sub", RemoveProductQuantityHandler).Methods("POST")
	router.HandleFunc("/products", GetAllProductsHandler).Methods("GET")

	// 启动服务器
	log.Fatal(http.ListenAndServe(":8080", handler))
	log.Println("RESTful server listening on port 8080...")
	//log.Fatal(http.ListenAndServe(":8080", router))
}
