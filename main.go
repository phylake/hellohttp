// Copyright 2019 Brandon Cook
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	random       string
	sizeResponse []byte
)

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

// curl -H "X-Req-URL: http://example.com" localhost:3000/client
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

// curl localhost:3000/size?byte_size=1024
func SizeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)

	byteSizeStr := r.URL.Query().Get("byte_size")
	if byteSizeStr == "" {
		w.WriteHeader(400)
		w.Write([]byte("missing byte_size query var"))
	}

	if byteSizeStr == os.Getenv("SIZE_RESPONSE_LEN") {
		w.Write(sizeResponse)
	} else {
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
}

// curl localhost:3000/delay?duration=1m
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

// curl -H "X-Filename: foo.bar" --data-binary @foo.bar localhost:3000/exfil
func ExfilHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	filename := r.Header.Get("X-Filename")
	if filename == "" {
		w.WriteHeader(400)
		return
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("os.OpenFile", err)
		w.WriteHeader(500)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		fmt.Println("os.OpenFile", err)
		w.WriteHeader(500)
		return
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	w.WriteHeader(404)
}

var healthy bool = true

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	if healthy {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
	}
}

// curl localhost:3000/health/pass
func HealthPassHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	healthy = true
}

// curl localhost:3000/health/fail
func HealthFailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	healthy = false
}

// polls for 1 hour
// iterations and sleep_duration are optional.
// not setting them is equivalent to:
// curl localhost:3000/longpoll?sleep_duration=30s&iterations=120
func LongPollHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	f := w.(http.Flusher)

	sleepDurationStr := r.URL.Query().Get("sleep_duration")
	if sleepDurationStr == "" {
		sleepDurationStr = "30s"
	}

	iterationsStr := r.URL.Query().Get("iterations")
	if iterationsStr == "" {
		iterationsStr = "120"
	}

	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	sleepDuration, err := time.ParseDuration(sleepDurationStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	for i := 0; i < iterations; i++ {
		w.Write([]byte("not dead yet\n"))
		f.Flush()
		time.Sleep(sleepDuration)
	}
}

// stream 1MiB of the alphabet by default
// chunk_size and chunk_count are optional.
// not setting them is equivalent to:
// curl localhost:3000/stream?chunk_size=1024&chunk_count=1024
func StreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-HelloHttp-Instance", random)
	f := w.(http.Flusher)

	chunkSizeStr := r.URL.Query().Get("chunk_size")
	if chunkSizeStr == "" {
		chunkSizeStr = "1024"
	}

	chunkCountStr := r.URL.Query().Get("chunk_count")
	if chunkCountStr == "" {
		chunkCountStr = "1024"
	}

	chunkCount, err := strconv.Atoi(chunkCountStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	chunkSize, err := strconv.Atoi(chunkSizeStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	for i := 0; i < chunkCount; i++ {
		bs := make([]byte, chunkSize)
		for i := 0; i < chunkSize; i++ {
			bs[i] = byte(97 + i%26)
		}
		w.Write(bs)
		f.Flush()
	}
}

func init() {
	bs := make([]byte, 4)
	rand.Read(bs)
	random = hex.EncodeToString(bs)

	if byteSize, err := strconv.Atoi(os.Getenv("SIZE_RESPONSE_LEN")); err == nil {
		sizeResponse = make([]byte, byteSize)
		for i := 0; i < byteSize; i++ {
			sizeResponse[i] = byte(97 + i%26)
		}
	}
}

func main() {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	server := &http.Server{
		Addr: ":" + port,
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
	http.DefaultServeMux.HandleFunc("/exfil", ExfilHandler)
	http.DefaultServeMux.HandleFunc("/404", NotFoundHandler)
	http.DefaultServeMux.HandleFunc("/health", HealthHandler)
	http.DefaultServeMux.HandleFunc("/health/pass", HealthPassHandler)
	http.DefaultServeMux.HandleFunc("/health/fail", HealthFailHandler)
	http.DefaultServeMux.HandleFunc("/longpoll", LongPollHandler)
	http.DefaultServeMux.HandleFunc("/stream", StreamHandler)

	fmt.Println("listening on", port)
	server.ListenAndServe()
}
