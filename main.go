package main

import (
	"fmt"
	"net/http"
	"os"
	"git.ventos.tk/gydos/pacmon/udp"
)

func handler(w http.ResponseWriter, r *http.Request) {

	var packagePath = "/var/cache/pacman/pkg" + r.URL.String()

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		fmt.Println("Fehla!")
		http.NotFound(w, r)
	} else {
		fmt.Println("Package found!")
		http.ServeFile(w, r, packagePath)
	}


	fmt.Println(r.URL)
}

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "setup" {
			fmt.Println("Setup found!")
			udp.SendMulicast("here?")
		} else {
			server()
		}
	} else {
		server()
	}

}

func server() {
	go udp.ServeMulticastUDP(udp.DebugHandler)
	http.HandleFunc("/", handler)
  	http.ListenAndServe(":8080", nil)
}
