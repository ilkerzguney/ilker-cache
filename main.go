package main

import (
	"math/rand"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"
)
var peers string
var port string

func main() {
	flag.StringVar(&port, "port", ":8081", "HTTP server port")
	flag.StringVar(&peers, "peers", "", "Comma seperated list of peer addresses")
	flag.Parse()

	nodeID := fmt.Sprintf("%s%d", "node", rand.Intn(100))
	peerList := strings.Split(peers, ",")

	cs := NewCacheServer(peerList, nodeID)
	cs.cache.starEvictionTicker(1 * time.Minute)

	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)
	fmt.Println("Server started at port:", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
