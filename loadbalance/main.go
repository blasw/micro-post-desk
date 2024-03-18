package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Server struct {
	url  string
	load float64
}

var LeastLoadedInstance Server = Server{load: 1}

// Specifies which microservice this loadbalancer belongs to
var INSTANCE = os.Getenv("INSTANCE")

// Should run in a parallel. Accepts: cancellation bool channel, array of servers names in network. Runs an infinite loop which finds the least loaded server and modifies global LeastLoadedInstance struct. Can be closed using cancellation channel.
func getLeastLoadedServer(cancel chan bool, servers []string) {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	for {
		select {
		case <-cancel:
			return

		default:
			for _, server := range servers {
				wg.Add(1)
				go func(server string) {
					defer wg.Done()

					resp, err := http.Get(fmt.Sprintf("http://%v/%v/load", server, INSTANCE))
					if err != nil {
						fmt.Println("Unable to get load:", err)
						return
					}

					defer resp.Body.Close()

					b, err := io.ReadAll(resp.Body)
					if err != nil {
						fmt.Println("Unable to read response's body:", err)
						return
					}
					loadStr := strings.Trim(string(b), "[]")

					load, err := strconv.ParseFloat(loadStr, 64)
					if err != nil {
						fmt.Println("Unable to parse float value: ", err)
						return
					}

					if LeastLoadedInstance.load > load {
						mutex.Lock()
						LeastLoadedInstance.load = load
						LeastLoadedInstance.url = server
						mutex.Unlock()
					}
				}(server)
			}
			wg.Wait()
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Unable to load servers from .env")
	}
	serv1 := os.Getenv("SERVER1")
	serv2 := os.Getenv("SERVER2")

	if serv1 == "" || serv2 == "" {
		log.Fatal("Unable to load servers from environment")
	}

	cancel := make(chan bool)

	r := gin.Default()

	servers := []string{serv1, serv2}

	go getLeastLoadedServer(cancel, servers)

	r.Any("/*path", func(c *gin.Context) {
		target, err := url.Parse(fmt.Sprintf("http://%v", LeastLoadedInstance.url))
		if err != nil {
			fmt.Println("Unable to parse target url: ", err)
			return
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(target)
		c.Request.URL.Path = c.Param("path")

		println(c.Request.URL.Path)

		reverseProxy.ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":5000")
}
