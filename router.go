package main

import (
	"net/http"
	"strings"
)

type router struct {
	// 키: http 메서드
	// 값: URL 패터별로 실행할 HandlerFunc
	handlers map[string]map[string]HandlerFunc
}

func (r *router) HandlerFunc(method, pattern string, h HandlerFunc) {
	m, ok := r.handlers[method]

	if !ok {
		m = make(map[string]HandlerFunc)
		r.handlers[method] = m
	}

	m[pattern] = h
}

func (r *router) handler() HandlerFunc {
	return func(c *Context) {
		for pattern, handler := range r.handlers[c.Request.Method] {
			if ok, params := match(pattern, c.Request.URL.Path); ok {
				for k, v := range params {
					c.Params[k] = v
				}

				handler(c)
				return
			}
		}

		http.NotFound(c.ResponseWriter, c.Request)
		return
	}
}

func match(pattern, path string) (bool, map[string]string) {
	// 패턴과 패스가 정확히 일치하면 바로 true 반환
	if pattern == path {
		return true, nil
	}

	// 패턴과 패스를 "/" 단위로 구분
	patterns := strings.Split(pattern, "/")
	paths := strings.Split(path, "/")

	// 패턴과 패스를 "/"로 구분한 후 부분 문자열 집합의 개수가 다르면 false 반환
	if len(patterns) != len(paths) {
		return false, nil
	}

	// 패턴이 일치하는 URL 매개변수를 담기 위한 params 맵 생성
	params := make(map[string]string)

	// "/"로 구분된 패턴/패스의 각 문자열을 하나씩 비교
	for i := 0; i < len(patterns); i++ {
		switch {
		case patterns[i] == paths[i]:
			continue
		case len(patterns[i]) > 0 && patterns[i][0] == ':':
			params[patterns[i][1:]] = paths[i]
		default:
			return false, nil
		}
	}

	return true, params
}
