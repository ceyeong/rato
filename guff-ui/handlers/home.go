package handlers

import (
	"context"
	"encoding/json"
	"guff-ui/pb"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Home : homepage
func (w *Web) Home(c echo.Context) error {
	data := echo.Map{"csrf_token": c.Get("csrf"), "user": c.Get("user")}
	if c.Request().Method == "POST" && c.FormValue("post") != "" {
		postID := c.FormValue("post")
		delReq := &pb.DeleteRequest{Uuid: postID}
		_, err := w.Client.Delete(context.TODO(), delReq)
		if err != nil {
			c.Logger().Error("delete failed. %v", err)
		}
		w.Cache.Delete("getAll")
	}
	cachedData, err := w.Cache.Get("getAll")
	if err != nil {
		print(err.Error())
	}
	if err == nil {
		m := new([]pb.Post)
		err = json.Unmarshal(cachedData, m)
		if err != nil {
			print(err.Error())
		}
		data["posts"] = m
		if err == nil {
			return c.Render(http.StatusOK, "home.html", data)
		}
	}
	req := &pb.GetAllRequest{}
	res, err := w.Client.GetAll(context.TODO(), req)
	if err != nil {
		return err
	}
	data["posts"] = res.Posts
	b, err := json.Marshal(res.Posts)
	if err == nil {
		w.Cache.Set("getAll", b)
	}
	return c.Render(http.StatusOK, "home.html", data)
}
