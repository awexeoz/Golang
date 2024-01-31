// handlers.go
package main

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Laptop представляет структуру для данных о ноутбуке
type Laptop struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Price int                `json:"price" bson:"price"`
}

func getLaptops(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "request",
		"status": "success",
	}).Info("GET request for laptops processed successfully")

	filter := r.URL.Query().Get("filter")
	sort := r.URL.Query().Get("sort")
	page := r.URL.Query().Get("page")

	limit := 10
	offset := 0

	// Расчет смещения на основе пагинации
	if p, err := strconv.Atoi(page); err == nil && p > 1 {
		offset = (p - 1) * limit
	}

	options := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)) // Опции для запроса к базе данных

	filterBson := bson.M{}
	if filter != "" {
		filterBson = bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: filter, Options: "i"}}} // Условие фильтрации
	}

	sortFields := strings.Split(sort, ",")
	for _, field := range sortFields {
		if field != "" {
			sortOrder := 1
			if strings.HasPrefix(field, "-") {
				sortOrder = -1
				field = strings.TrimPrefix(field, "-")
			}
			options = options.SetSort(map[string]int{field: sortOrder})
		}
	}

	// Обработка ошибок
	defer func() {
		if err := recover(); err != nil {
			log.WithField("error", err).Error("Internal Server Error")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}
	}()

	cur, err := collection.Find(context.Background(), filterBson, options) // Запрос к базе данных
	if err != nil {
		// Используйте logrus для логирования
		logrus.WithField("error", err).Error("Error querying the database")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	defer cur.Close(context.Background())

	var laptops []Laptop
	for cur.Next(context.Background()) {
		var laptop Laptop
		err := cur.Decode(&laptop)
		if err != nil {
			logrus.Error(err)
			continue
		}
		laptops = append(laptops, laptop)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(laptops) // Отправка данных в формате JSON клиенту
}
