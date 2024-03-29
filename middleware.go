package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type Middleware func(next HandlerFunc) HandlerFunc

// 로그 처리 미들웨어
func logHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		t := time.Now()

		next(c)

		// 웹 요청 정보와 전체 소요 시간을 로그로 남김
		log.Printf("[%s] %q %v\n",
			c.Request.Method,
			c.Request.URL.String(),
			time.Now().Sub(t))
	}
}

// 에러 처리 미들웨어
func recoverHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic %+v", err)

				http.Error(c.ResponseWriter,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}
		}()

		next(c)
	}
}

// 웹 요청 내용 파싱 미들웨어
func parseFormHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		c.Request.ParseForm()
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				c.Params[k] = v[0]
			}
		}

		next(c)
	}
}

func parseJsonBodyHandler(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		var m map[string]interface{}
		if json.NewDecoder(c.Request.Body).Decode(&m); len(m) > 0 {
			for k, v := range m {
				c.Params[k] = v
			}
		}

		next(c)
	}
}

// 정적 파일 내용을 전달하는 미들웨어
func staticHandler(next HandlerFunc) HandlerFunc {
	var (
		dir       = http.Dir('.')
		indexFile = "index.html"
	)

	return func(c *Context) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			next(c)
			return
		}

		file := c.Request.URL.Path
		f, err := dir.Open(file)
		if err != nil {
			next(c)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			next(c)
			return
		}

		// URL 경로가 디렉토리면 indexFile을 사용
		if fi.IsDir() {
			// 경로 끝이 "/" 아니라면
			if !strings.HasSuffix(c.Request.URL.Path, "/") {
				http.Redirect(c.ResponseWriter, c.Request, c.Request.URL.Path+"/", http.StatusFound)
				return
			}

			file = path.Join(file, indexFile)

			f, err := dir.Open(file)
			if err != nil {
				next(c)
				return
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil || fi.IsDir() {
				next(c)
				return
			}

			http.ServeContent(c.ResponseWriter, c.Request, file, fi.ModTime(), f)
			return
		}

		http.ServeContent(c.ResponseWriter, c.Request, file, fi.ModTime(), f)
	}
}
