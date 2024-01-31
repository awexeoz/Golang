package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var database *mongo.Database
var collection *mongo.Collection

// initMongoDB устанавливает соединение с базой данных MongoDB
func initMongoDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
		return
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Не удалось протестировать подключение к MongoDB:", err)
		return
	}

	database = client.Database("laptop_store")
	collection = database.Collection("laptops")

	fmt.Println("Успешное подключение к MongoDB")
}
