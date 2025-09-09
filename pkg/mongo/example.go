package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 用户模型示例
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Age       int                `bson:"age" json:"age"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Product 产品模型示例
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Category    string             `bson:"category" json:"category"`
	Stock       int                `bson:"stock" json:"stock"`
	Tags        []string           `bson:"tags" json:"tags"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ExampleBasicUsage 基础使用示例
func ExampleBasicUsage() {
	fmt.Println("=== MongoDB基础使用示例 ===")

	// 1. 创建客户端
	client, err := NewClientWithURI("mongodb://localhost:27017", "example_db")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	// 2. 创建仓储
	userRepo := NewRepository(client, "example_db", "users")

	// 3. 插入文档
	user := &User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()
	result, err := userRepo.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return
	}
	fmt.Printf("Inserted user with ID: %v\n", result.InsertedID)

	// 4. 查询文档
	var foundUser User
	err = userRepo.FindOne(ctx, bson.M{"email": "zhangsan@example.com"}).Decode(&foundUser)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// 5. 更新文档
	update := bson.M{"$set": bson.M{"age": 26, "updated_at": time.Now()}}
	updateResult, err := userRepo.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, update)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return
	}
	fmt.Printf("Updated %d document(s)\n", updateResult.ModifiedCount)

	// 6. 删除文档
	deleteResult, err := userRepo.DeleteOne(ctx, bson.M{"_id": foundUser.ID})
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return
	}
	fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)
}

// ExampleAdvancedQueries 高级查询示例
func ExampleAdvancedQueries() {
	fmt.Println("=== MongoDB高级查询示例 ===")

	client, err := NewClientWithURI("mongodb://localhost:27017", "example_db")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	productRepo := NewRepository(client, "example_db", "products")
	ctx := context.Background()

	// 1. 批量插入示例数据
	products := []interface{}{
		&Product{
			Name:        "笔记本电脑",
			Description: "高性能笔记本电脑",
			Price:       5999.99,
			Category:    "电子产品",
			Stock:       10,
			Tags:        []string{"电脑", "办公", "游戏"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		&Product{
			Name:        "智能手机",
			Description: "最新款智能手机",
			Price:       3999.99,
			Category:    "电子产品",
			Stock:       20,
			Tags:        []string{"手机", "通讯", "拍照"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		&Product{
			Name:        "办公椅",
			Description: "人体工学办公椅",
			Price:       899.99,
			Category:    "家具",
			Stock:       5,
			Tags:        []string{"椅子", "办公", "舒适"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	_, err = productRepo.InsertMany(ctx, products)
	if err != nil {
		log.Printf("Failed to insert products: %v", err)
		return
	}
	fmt.Println("Inserted sample products")

	// 2. 复杂查询
	filter := bson.M{
		"category": "电子产品",
		"price": bson.M{
			"$gte": 3000,
			"$lte": 6000,
		},
	}

	opts := options.Find().SetSort(bson.M{"price": -1}).SetLimit(10)
	var results []Product
	err = productRepo.FindAll(ctx, filter, &results, opts)
	if err != nil {
		log.Printf("Failed to find products: %v", err)
		return
	}

	fmt.Printf("Found %d products:\n", len(results))
	for _, product := range results {
		fmt.Printf("- %s: ¥%.2f\n", product.Name, product.Price)
	}

	// 3. 聚合查询
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":        "$category",
			"count":      bson.M{"$sum": 1},
			"avgPrice":   bson.M{"$avg": "$price"},
			"totalStock": bson.M{"$sum": "$stock"},
		}},
		{"$sort": bson.M{"avgPrice": -1}},
	}

	var aggregateResults []bson.M
	err = productRepo.AggregateAll(ctx, pipeline, &aggregateResults)
	if err != nil {
		log.Printf("Failed to aggregate: %v", err)
		return
	}

	fmt.Println("\nCategory statistics:")
	for _, result := range aggregateResults {
		fmt.Printf("Category: %s, Count: %v, Avg Price: ¥%.2f, Total Stock: %v\n",
			result["_id"], result["count"], result["avgPrice"], result["totalStock"])
	}

	// 4. 文本搜索（需要创建文本索引）
	textFilter := bson.M{"$text": bson.M{"$search": "电脑"}}
	var textResults []Product
	err = productRepo.FindAll(ctx, textFilter, &textResults)
	if err != nil {
		// 文本搜索可能失败（如果没有文本索引），这是正常的
		fmt.Printf("Text search failed (may need text index): %v\n", err)
	} else {
		fmt.Printf("\nText search results: %d products\n", len(textResults))
	}
}

// ExampleTransactions 事务示例
func ExampleTransactions() {
	fmt.Println("=== MongoDB事务示例 ===")

	client, err := NewClientWithURI("mongodb://localhost:27017", "example_db")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	userRepo := NewRepository(client, "example_db", "users")
	productRepo := NewRepository(client, "example_db", "products")
	tm := NewTransactionManager(client)

	ctx := context.Background()

	// 事务示例：转账操作
	err = tm.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		// 在事务中执行多个操作
		user1 := &User{
			Name:      "用户1",
			Email:     "user1@example.com",
			Age:       30,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		user2 := &User{
			Name:      "用户2",
			Email:     "user2@example.com",
			Age:       25,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 插入两个用户
		_, err = userRepo.InsertOne(sc, user1)
		if err != nil {
			return nil, fmt.Errorf("failed to insert user1: %w", err)
		}

		_, err = userRepo.InsertOne(sc, user2)
		if err != nil {
			return nil, fmt.Errorf("failed to insert user2: %w", err)
		}

		// 插入一个产品
		product := &Product{
			Name:        "事务测试产品",
			Description: "在事务中创建的产品",
			Price:       199.99,
			Category:    "测试",
			Stock:       1,
			Tags:        []string{"测试", "事务"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err = productRepo.InsertOne(sc, product)
		if err != nil {
			return nil, fmt.Errorf("failed to insert product: %w", err)
		}

		fmt.Println("Transaction completed successfully")
		return nil, nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}

	fmt.Println("All operations completed in transaction")
}

// ExampleValidation 验证示例
func ExampleValidation() {
	fmt.Println("=== MongoDB验证示例 ===")

	client, err := NewClientWithURI("mongodb://localhost:27017", "example_db")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	// 创建用户Schema
	userSchema := map[string]interface{}{
		"type":     "object",
		"required": []interface{}{"name", "email", "age"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":      "string",
				"minLength": 2,
				"maxLength": 50,
			},
			"email": map[string]interface{}{
				"type":      "string",
				"minLength": 5,
				"maxLength": 100,
			},
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"maximum": 150,
			},
			"status": map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"active", "inactive", "pending"},
			},
		},
	}

	// 创建验证器
	validator := NewSchemaValidator(userSchema, client.GetLogger())

	// 创建带验证的仓储
	userRepo := NewRepository(client, "example_db", "users")
	validatedRepo := NewValidatedRepository(userRepo, validator)

	ctx := context.Background()

	// 1. 有效用户（应该成功）
	validUser := &User{
		Name:      "李四",
		Email:     "lisi@example.com",
		Age:       28,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = validatedRepo.InsertOne(ctx, validUser)
	if err != nil {
		log.Printf("Failed to insert valid user: %v", err)
	} else {
		fmt.Println("Successfully inserted valid user")
	}

	// 2. 无效用户（应该失败）
	invalidUser := &User{
		Name:      "王",             // 太短
		Email:     "invalid-email", // 无效邮箱格式
		Age:       200,             // 超出范围
		Status:    "unknown",       // 不在枚举中
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = validatedRepo.InsertOne(ctx, invalidUser)
	if err != nil {
		fmt.Printf("Validation failed as expected: %v\n", err)
	} else {
		fmt.Println("Unexpected: invalid user was inserted")
	}

	// 3. 验证更新操作
	validUpdate := bson.M{"$set": bson.M{"age": 30}}
	_, err = validatedRepo.UpdateOne(ctx, bson.M{"email": "lisi@example.com"}, validUpdate)
	if err != nil {
		log.Printf("Failed to update: %v", err)
	} else {
		fmt.Println("Successfully updated user")
	}

	// 4. 无效更新（应该失败）
	invalidUpdate := bson.M{"$set": bson.M{"age": "not a number"}}
	_, err = validatedRepo.UpdateOne(ctx, bson.M{"email": "lisi@example.com"}, invalidUpdate)
	if err != nil {
		fmt.Printf("Update validation failed as expected: %v\n", err)
	} else {
		fmt.Println("Unexpected: invalid update was applied")
	}
}

// ExampleCustomConfiguration 自定义配置示例
func ExampleCustomConfiguration() {
	fmt.Println("=== MongoDB自定义配置示例 ===")

	// 创建自定义配置
	config := &Config{
		URI:      "mongodb://localhost:27017",
		Database: "custom_db",
		Pool: PoolConfig{
			MaxPoolSize:     50,
			MinPoolSize:     5,
			MaxConnIdleTime: 10 * time.Minute,
		},
		Log: LogConfig{
			Enabled:            true,
			Level:              "debug",
			SlowQuery:          true,
			SlowQueryThreshold: 50 * time.Millisecond,
		},
		ReadWrite: ReadWriteConfig{
			ReadPreference: "secondaryPreferred",
			WriteConcern:   "majority",
			ReadConcern:    "local",
		},
		Timeout: TimeoutConfig{
			Connect:         5 * time.Second,
			ServerSelection: 10 * time.Second,
			Socket:          30 * time.Second,
			Heartbeat:       5 * time.Second,
		},
	}

	// 使用自定义配置创建客户端
	client, err := NewClient(config)
	if err != nil {
		log.Printf("Failed to create client with custom config: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("Successfully created client with custom configuration")

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}

	fmt.Println("Health check passed")

	// 获取统计信息
	stats, err := client.Stats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return
	}

	fmt.Printf("Server stats: host=%v, version=%v\n",
		stats["host"], stats["version"])
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	fmt.Println("开始运行MongoDB模块示例...")
	fmt.Println()

	// 注意：这些示例需要运行的MongoDB实例
	// 在生产环境中使用前，请确保MongoDB服务正在运行

	ExampleBasicUsage()
	fmt.Println()

	ExampleAdvancedQueries()
	fmt.Println()

	ExampleTransactions()
	fmt.Println()

	ExampleValidation()
	fmt.Println()

	ExampleCustomConfiguration()
	fmt.Println()

	fmt.Println("所有示例运行完成！")
}
