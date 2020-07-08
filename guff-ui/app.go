package main

import (
	"guff-ui/handlers"
	"guff-ui/pb"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

// Template : Implements echo.Renderer
type Template struct {
	templates *template.Template
}

// Render :
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	os.Setenv("SESSION_SECRET", "sjdksfkbi333590329dme900002")
	os.Setenv("GRPC_HOST", "127.0.0.1:8080")
	os.Setenv("SERVE_PORT", "8081")
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(csrf())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))

	connection := connectGrpc()
	defer connection.Close()
	client := pb.NewPostServiceClient(connection)

	//cache
	apiCache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Hour))
	if err != nil {
		e.Logger.Fatalf("%v", err)
	}
	webCache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Hour))
	if err != nil {
		e.Logger.Fatalf("%v", err)
	}
	web := handlers.Web{Client: client, Cache: webCache}
	api := handlers.API{Client: client, Cache: apiCache}

	t := &Template{
		templates: template.Must(template.ParseGlob("public/*.html")),
	}
	e.Renderer = t
	e.Use(sessionAuth())
	// Web
	e.GET("/create", web.CreatePost)
	e.POST("/create", web.CreatePost)

	e.GET("/login", web.Login)
	e.POST("/login", web.Login)
	e.GET("/", web.Home)
	e.POST("/", web.Home)

	e.GET("/logout", web.Logout)
	e.POST("/logout", web.Logout)

	//Api
	g := e.Group("/api/v1")
	g.GET("/getAll", api.GetAll)

	e.Logger.Fatal(e.Start(":" + os.Getenv("SERVE_PORT")))
}

func csrf() echo.MiddlewareFunc {
	config := middleware.DefaultCSRFConfig
	config.TokenLookup = "form:csrf_token"
	return middleware.CSRFWithConfig(config)
}

func sessionAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api") {
				return next(c)
			}
			sess, _ := session.Get("session", c)
			c.Set("user", sess.Values["user"])
			if c.Request().URL.Path == "/login" || c.Request().URL.Path == "/" {
				return next(c)
			}
			if sess.Values["user"] == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}

func connectGrpc() *grpc.ClientConn {
	connection, err := grpc.Dial(os.Getenv("GRPC_HOST"), grpc.WithInsecure())
	if err != nil {
		log.Fatal("grpc connection failed", err)
	}
	return connection
}
