package utils

import (
	"log"
	"os"
)

func CheckENVS() {
	error_messages := []string{}

	if os.Getenv("DB_ADDR") == "" {
		error_messages = append(error_messages, "DB_ADDR ENV is not set")
	}

	if os.Getenv("CACHE_ADDR") == "" {
		error_messages = append(error_messages, "CACHE_ADDR ENV is not set")
	}

	if os.Getenv("USERS_LOADBALANCER") == "" {
		error_messages = append(error_messages, "USERS_LOADBALANCER ENV is not set")
	}

  if len(error_messages) == 0 {
    return
  }

	for _, v := range error_messages {
    log.Println(v)
	}

	log.Fatal("Unable to start service without ENVs")
}
