package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func setupRouter(config Config) *Router {
	router := newRouter()
	router.GET("/", handleRoot())
	router.GET("/meep", handleGetMeep())
	router.GET("/:meep/mino/:mino", handleDoubleWild())
	router.POST("/files/:filename", handlePostFile(config.publicDir))
	router.GET("/files/:filename", handleGetFile(config.publicDir))
	router.GET("/echo/:text", handleEchoText())
	router.GET("/user-agent", handleEchoUserAgent())
	return &router
}

func handleRoot() Handler {
	return func(req *Request, res *ResponseWriter) {
		res.WriteHeader(http.StatusOK)
	}
}

func handleGetMeep() Handler {
	return func(req *Request, res *ResponseWriter) {
		res.WriteHeader(http.StatusOK)
	}
}

func handleDoubleWild() Handler {
	return func(req *Request, res *ResponseWriter) {
		res.WriteHeader(http.StatusOK)
	}
}

func handlePostFile(publicDir string) Handler {
	return func(req *Request, res *ResponseWriter) {
		filename, _ := req.params["filename"]
		fp := filepath.Join(publicDir, filename)

		sizeStr, _ := req.headers["Content-Length"]
		size, _ := strconv.Atoi(sizeStr)
		contents := make([]byte, size)
		_, err := io.ReadFull(req.body, contents)
		if err != nil {
			fmt.Printf("err reading request body %s\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(fp, contents, 0644)
		if err != nil {
			fmt.Printf("err writing to file: %s", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusCreated)
	}
}

func handleGetFile(publicDir string) Handler {
	return func(req *Request, res *ResponseWriter) {
		filename, _ := req.params["filename"]

		fp := filepath.Join(publicDir, filename)
		_, err := os.Stat(fp)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				res.WriteHeader(http.StatusNotFound)
				return
			}
			res.WriteHeader(http.StatusInternalServerError)
		}

		data, err := os.ReadFile(fp)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Status(http.StatusOK)
		res.Headers["Content-Type"] = "application/octet-stream"
		res.Headers["Content-Length"] = fmt.Sprintf("%d", len(data))
		res.Write(data)
	}
}

func handleEchoText() Handler {
	return func(req *Request, res *ResponseWriter) {
		text, _ := req.params["text"]

		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = fmt.Sprintf("%d", len(text))

		encodingsHeader, ok := req.headers["Accept-Encoding"]
		if !ok {
			res.Status(http.StatusOK)
			res.Write([]byte(text))
			return
		}

		encodings := strings.Split(strings.ReplaceAll(encodingsHeader, " ", ""), ",")
		if len(encodings) == 0 {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var allowGzip bool
		for _, encoding := range encodings {
			if encoding == "gzip" {
				allowGzip = true
				break
			}
		}

		if !allowGzip {
			res.Status(http.StatusOK)
			res.Write([]byte(text))
			return
		}

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		_, err := gzipWriter.Write([]byte(text))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = gzipWriter.Close()
		if err != nil {
			fmt.Printf("err closing gzipwriter %s", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println(len(text), buf.Len())

		data := buf.Bytes()
		res.Headers["Content-Encoding"] = "gzip"
		res.Headers["Content-Length"] = fmt.Sprintf("%d", len(data))
		res.Status(http.StatusOK)
		n, err := res.Write(data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(n)
	}
}

func handleEchoUserAgent() Handler {
	return func(req *Request, res *ResponseWriter) {
		userAgent, ok := req.headers["User-Agent"]
		if !ok {
			fmt.Println("Failed to find value for param 'text'")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Status(http.StatusOK)
		res.Headers["Content-Type"] = "text/plain"
		res.Headers["Content-Length"] = fmt.Sprintf("%d", len(userAgent))
		res.Write([]byte(userAgent))
	}
}
