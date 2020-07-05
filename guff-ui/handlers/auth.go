package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Login : login
func (w *Web) Login(c echo.Context) error {
	data := echo.Map{"csrf_token": c.Get("csrf")}
	if c.Request().Method == "GET" {

		c.Render(http.StatusOK, "login.html", data)
	}
	if c.Request().Method == "POST" {

		email := c.FormValue("email")
		password := c.FormValue("password")
		if !(email == os.Getenv("USER_EMAIL") && password == os.Getenv("USER_PASSWORD")) {
			data["error"] = "username or password doesn't match"
			return c.Render(http.StatusOK, "login.html", data)
		}
		sess, _ := session.Get("session", c)
		sess.Values["user"] = email
		sess.Save(c.Request(), c.Response())
		return c.Redirect(http.StatusFound, "/")
	}
	return nil
}

// Logout : log out
func (w *Web) Logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values = nil
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/login")
}
