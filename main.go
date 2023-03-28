package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type User struct {
	Id        string
	AddressId string
}

func main() {

	s := NewServer()

	s.HandlerFunc("GET", "/", func(c *Context) {
		c.RenderTemplate("/public/index.html", map[string]interface{}{"time": time.Now()})
	})

	s.HandlerFunc("GET", "/about", func(c *Context) {
		fmt.Fprintln(c.ResponseWriter, "about")
	})

	s.HandlerFunc("GET", "/users/:id", func(c *Context) {
		u := User{Id: c.Params["id"].(string)}
		c.RenderXml(u)
	})

	s.HandlerFunc("GET", "/users/:user_id/addresses/:address_id", func(c *Context) {
		u := User{Id: c.Params["user_id"].(string), AddressId: c.Params["address_id"].(string)}
		c.RenderJson(u)
	})

	s.HandlerFunc("POST", "/users", func(c *Context) {
		fmt.Fprintln(c.ResponseWriter, c.Params)
	})

	s.HandlerFunc("POST", "/users/:user_id/addresses", func(c *Context) {
		fmt.Fprintf(c.ResponseWriter, "create user %v's address\n", c.Params["user_id"])
	})

	s.HandlerFunc("GET", "/public", func(c *Context) {
		fmt.Fprintf(c.ResponseWriter, "public")
	})

	s.HandlerFunc("GET", "/login", func(c *Context) {
		c.RenderTemplate("/public/login.html", map[string]interface{}{"message": "로그인이 필요합니다."})
	})

	s.HandlerFunc("POST", "/login", func(c *Context) {
		fmt.Println("dd")
		if CheckLogin(c.Params["username"].(string), c.Params["password"].(string)) {
			http.SetCookie(c.ResponseWriter, &http.Cookie{
				Name:  "X_AUTH",
				Value: Sign(VerifyMessage),
				Path:  "/",
			})
			c.Redirect("/")
		}

		c.RenderTemplate("/public/login.html", map[string]interface{}{"message": "id와 password를 확인해주세요."})
	})

	s.Use(AuthHandler)

	s.Run(":7711")
}

func CheckLogin(username, password string) bool {
	const (
		USERNAME = "tester"
		PASSWORD = "1234"
	)

	return username == USERNAME && password == PASSWORD
}

const VerifyMessage = "verified"

func AuthHandler(next HandlerFunc) HandlerFunc {
	ignore := []string{"/login", "public/index.html"}

	return func(c *Context) {
		// URL prefix가 ignore에 있다면 auth를 체크하지 않음
		for _, s := range ignore {
			if strings.HasPrefix(c.Request.URL.Path, s) {
				next(c)
				return
			}
		}

		if v, err := c.Request.Cookie("X_AUTH"); err == http.ErrNoCookie {
			c.Redirect("/login")
			return

		} else if err != nil {
			c.RenderErr(http.StatusInternalServerError, err)
			return

		} else if Verify(VerifyMessage, v.Value) {
			// 인증 완료
			next(c)
			return
		}

		c.Redirect("/login")
	}
}

func Sign(message string) string {
	secretKey := []byte("ttest")
	if len(secretKey) == 0 {
		return ""
	}

	mac := hmac.New(sha1.New, secretKey)
	io.WriteString(mac, message)
	return hex.EncodeToString(mac.Sum(nil))
}

// 인증 토큰 확인
func Verify(message, sig string) bool {
	return hmac.Equal([]byte(sig), []byte(Sign(message)))
}
