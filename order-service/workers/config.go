package main

import "os"

type Config struct {
	KafkaBrokers string
	GroupID      string
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
}

func LoadConfig() Config {
	return Config{
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
		GroupID:      os.Getenv("KAFKA_GROUP_ID"),
		DBHost:       os.Getenv("POSTGRES_HOST"),
		DBUser:       os.Getenv("POSTGRES_USER"),
		DBPassword:   os.Getenv("POSTGRES_PASSWORD"),
		DBName:       os.Getenv("POSTGRES_DB"),
	}
}
