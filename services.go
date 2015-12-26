package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"gopkg.in/redis.v3"
)

var services = struct {
	config Config
	db     *sqlx.DB
	redis  *redis.Client
}{}

func InitServices() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}
	services.config = config

	maxRetries := 4

	if err := retry("setupDB", maxRetries, setupDB); err != nil {
		return err
	}

	if err := retry("setupRedis", maxRetries, setupRedis); err != nil {
		return err
	}

	if err := retry("setupRabbitMQ", maxRetries, setupRabbitMQ); err != nil {
		return err
	}

	tokensStartDeleteExpiredLoop()

	return nil
}

func retry(label string, maxRetries int, fn func() error) error {

	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		} else {
			pow := math.Max(float64(i), 8)
			millis := 10 * math.Pow(2, pow) * rand.Float64()
			log.Printf("%v failed on try #%v, sleep %v millis", label, i, millis)
			time.Sleep(time.Duration(millis) * time.Millisecond)
		}
	}

	return err
}

func setupDB() error {
	pgcfg := services.config.Postgres
	url := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", pgcfg.User, pgcfg.Password, pgcfg.Addr, pgcfg.Database)
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return fmt.Errorf("unable to create db driver with params %v: %v", url, err)
	}

	_, err = db.Query("SELECT 1")
	if err != nil {
		return fmt.Errorf("unable to connect to db with params %v: %v", url, err)
	}

	services.db = db
	return nil
}

func setupRedis() error {
	client := redis.NewClient(&redis.Options{
		Addr:     services.config.RedisAddr,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return fmt.Errorf("unable to ping redis server at %v: %v", services.config.RedisAddr, err)
	}
	services.redis = client
	return nil
}

func setupRabbitMQ() error {
	url := fmt.Sprintf("amqp://%s", services.config.RabbitMQAddr)
	connection, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("rabbitmq: unable to connect to url %v: %v", url, err)
	}
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("rabbitmq: unable to open channel on connection to %v: %v", url, err)
	}

	channel.Confirm(false)
	// services.channel = channel
	//
	// q, err := channel.QueueDeclare("feedUpdates", true, false, false, false, nil)
	// if err != nil {
	// 	return fmt.Errorf("rabbitmq: error declaring queue: %v", err)
	// }

	return nil
}
