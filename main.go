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
	"io"
	"errors"
)

const (
	port = 8080
	UDP_SERVER_COMMAND = "GyDOS: PACMON: server:"
	UDP_DISCOVER_COMMAND = "GyDOS: PACMON: discover"
)

var (
	servers = make(map[string]bool)
)

func Proxy(w http.ResponseWriter, r *http.Request, server string) error {
	resp, err := http.Get(server + r.URL.String())
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		for h := range resp.Header {
			w.Header().Add(h, resp.Header.Get(h))
		}
		_, err := io.Copy(w, resp.Body)
		return err
	} else {
		return errors.New("status")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

	var packagePath = "/var/cache/pacman/pkg" + r.URL.String()

	fmt.Println(r.URL)

	ip := r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
	ip = strings.Trim(ip, "[]")
	if net.ParseIP(ip).IsLoopback() {
		for server, _ := range servers {
			err := Proxy(w, r, server)
			if err == nil {
				return
			} else {
				fmt.Println(err)
			}
		}
		http.NotFound(w, r)
		/*mirrorlist, err := os.Open("/etc/pacman.d/mirrorlist")
		if err != nil {
			log.Fatal(err)
		}
		bufR := bufio.NewReader(mirrorlist)
		for {
			l, _, err := bufR.ReadLine()
			if err != nil {
				continue
			}
			sl := string(l)
			if strings.HasPrefix(sl, "#") {
				continue
			}
			if strings.HasPrefix(sl, "Server = ") {
				Proxy(w, r, strings.TrimPrefix(sl, "Server = "))
			}
		}*/
	} else {
		if _, err := os.Stat(packagePath); os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.ServeFile(w, r, packagePath)
		}
	}
}

func main() {
	fmt.Println("Started, local IP:", udp.GetLocalIP())
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
			server := "http://" + message[len(UDP_SERVER_COMMAND) + 1:]
			//fmt.Println("Server found: ", server)
			servers[server] = true
		}
	}
}
