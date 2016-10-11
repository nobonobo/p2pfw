package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/rs/cors"
)

func getIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	ip, err := getIP()
	if err != nil {
		log.Fatalln("getip failed:", err)
	}
	l, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/", cors.Default().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remote := r.Header.Get("X-Forwarded-For")
		if remote == "" {
			if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				remote = h
			}
		}
		fmt.Fprint(w, remote)
	})))
	log.Println("signaling server:", l.Addr())
	go func() {
		if err := http.Serve(l, nil); err != nil {
			log.Fatalln(err)
		}
	}()
	cmd := exec.Command(
		"turnserver", "-n", "--log-file=stdout",
		fmt.Sprintf("--external-ip=%s", ip))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
