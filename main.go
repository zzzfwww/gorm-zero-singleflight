package main

import (
	"context"
	"gorm-test/server"
	"gorm-test/syncx"
	"log"
)

// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
var singleFlights = syncx.NewSingleFlight()

func main() {
	client := server.GormMysql(&server.MysqlConfig{
		Path:         "localhost",
		Password:     "yourpassword",
		Dbname:       "gorm",
		Username:     "root",
		Port:         "3306",
		MaxIdleConns: 8,
		MaxOpenConns: 8,
	})
	log.Println(client)
	ctx := context.Background()
	r := server.Redis(&server.RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	srv := server.NewGetService(client, r, singleFlights)
	info, err := srv.FindOne(ctx, 1)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println(info.TradeStateDesc)
}
