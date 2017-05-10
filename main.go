package main

import (
	"fmt"
	"net/http"
	"os"
	"git.ventos.tk/gydos/pacmon/udp"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	port = 8080
	UDP_SERVER_COMMAND = "GyDOS: PACMON: server:"
	UDP_DISCOVER_COMMAND = "GyDOS: PACMON: discover"
)

var (
	servers = make(map[string]bool)
)

func handler(w http.ResponseWriter, r *http.Request) {

	var packagePath = "/var/cache/pacman/pkg" + r.URL.String()

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		fmt.Println(r.RemoteAddr)
		http.NotFound(w, r)
	} else {
		http.ServeFile(w, r, packagePath)
	}

	fmt.Println(r.URL)
}

func main() {
	fmt.Println(udp.GetLocalIP())
	go udp.ServeMulticastUDP(UDPHandler)
	time.Sleep(time.Second)
	go udp.SendMulicast(UDP_DISCOVER_COMMAND)
	time.Sleep(time.Second * 3)
	server()
}

func server() {
	http.HandleFunc("/", handler)
  	http.ListenAndServe(":"+strconv.FormatInt(port, 10), nil)
}

func UDPHandler(src *net.UDPAddr, n int, b []byte) {
	var message = string(b[:n])

	if message == UDP_DISCOVER_COMMAND {
		udp.SendMulicast(UDP_SERVER_COMMAND + " " + udp.GetLocalIP() + ":" + strconv.FormatInt(port, 10))
	} else if strings.HasPrefix(message, UDP_SERVER_COMMAND) {
		if strings.Split(message[len(UDP_SERVER_COMMAND) + 1:], ":")[0] != udp.GetLocalIP() {
			server := "Server = http://" + message[len(UDP_SERVER_COMMAND) + 1:]
			servers[server] = true
		}
	}
}
