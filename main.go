package main

import (
	"fmt"
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

	s.HandlerFunc("GET", "/public", staticHandler(func(c *Context) {
		fmt.Fprintf(c.ResponseWriter, "public")
	}))

	s.Run(":7711")
}
