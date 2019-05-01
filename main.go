package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

var random string

func PongHandler(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("X-HelloHttp-Instance", random)
	w.Header().Set("X-HelloHttp-Req-Body-Length", strconv.Itoa(len(bs)))
	w.Write([]byte("PONG"))
}

func LogRequestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("")
	fmt.Println("Proto", r.Proto)
	fmt.Println("TransferEncoding", r.TransferEncoding)
	fmt.Println("Close", r.Close)
	fmt.Println("Host", r.Host)
	fmt.Println("RemoteAddr", r.RemoteAddr)
	for k, v := range r.Header {
		fmt.Println("Header", k, v)
	}

	w.Header().Set("X-HelloHttp-Instance", random)
	w.Write([]byte("PONG"))
}

func ClientHandler(w http.ResponseWriter, r *http.Request) {
	urlStr := r.Header.Get("X-Req-URL")
	if urlStr == "" {
		w.WriteHeader(400)
		w.Write([]byte("missing X-Req-URL"))
		return
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("X-HelloHttp-Instance", random)
	httputil.NewSingleHostReverseProxy(u).ServeHTTP(w, r)
}

func SizeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)

	byteSizeStr := r.URL.Query().Get("byte_size")
	if byteSizeStr == "" {
		w.WriteHeader(400)
		w.Write([]byte("missing byte_size query var"))
	}

	byteSize, err := strconv.Atoi(byteSizeStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("strconv.Atoi failed"))
	}

	bs := make([]byte, byteSize)
	for i := 0; i < byteSize; i++ {
		bs[i] = byte(97 + i%26)
	}
	w.Write(bs)
}

func DelayHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		w.WriteHeader(400)
		w.Write([]byte("missing duration query var"))
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("time.ParseDuration failed"))
	}

	time.Sleep(duration)
}

func init() {
	bs := make([]byte, 4)
	rand.Read(bs)
	random = hex.EncodeToString(bs)
}

func main() {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}

	server := &http.Server{
		Addr: ":3000",
	}

	if d, err := time.ParseDuration(os.Getenv("IDLE_TIMEOUT")); err != nil {
		server.IdleTimeout = d
	}

	http.DefaultServeMux.HandleFunc("/", PongHandler)
	http.DefaultServeMux.HandleFunc("/ping", PongHandler)
	http.DefaultServeMux.HandleFunc("/log", LogRequestHandler)
	http.DefaultServeMux.HandleFunc("/client", ClientHandler)
	http.DefaultServeMux.HandleFunc("/size", SizeHandler)
	http.DefaultServeMux.HandleFunc("/delay", DelayHandler)

	fmt.Println("listening on 3000")
	server.ListenAndServe()
}
