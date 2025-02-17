package mongodb

import (
	"context"
	"fmt"
	"strings"

	"QA-System/internal/global/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// QA mongodb存储答卷的集合名
var QA string

// Record mongodb存储记录的集合名
var Record string

// Init 初始化 MongoDB 连接
func Init() *mongo.Database {
	// Get MongoDB connection information from the configuration file
	user := config.Config.GetString("mongodb.user")
	pass := config.Config.GetString("mongodb.pass")
	host := config.Config.GetString("mongodb.host")
	port := config.Config.GetString("mongodb.port")
	db := config.Config.GetString("mongodb.db")
	QA = config.Config.GetString("mongodb.qa-collection")
	Record = config.Config.GetString("mongodb.record-collection")

	var dsn string
	if strings.Contains(host, "mongodb.net") {
		// MongoDB Atlas 连接字符串
		dsn = fmt.Sprintf("mongodb+srv://%v:%v@%v/%v?retryWrites=true&w=majority&appName=jhQAtest",
			user, pass, host, db)
	} else {
		// 普通 MongoDB 连接字符串
		dsn = fmt.Sprintf("mongodb://%v:%v@%v:%v/%v", user, pass, host, port, db)
	}

	// 使用 dsn 连接 MongoDB
	clientOptions := options.Client().ApplyURI(dsn)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		zap.L().Fatal("Failed to connect to MongoDB:" + err.Error())
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		zap.L().Fatal("Failed to ping MongoDB:" + err.Error())
	}

	mdb := client.Database(db)

	// 日志记录
	zap.L().Info("Connected to MongoDB")
	return mdb
}
