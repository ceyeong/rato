package handlers

import (
	"context"
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
	}
	req := &pb.GetAllRequest{}
	res, err := w.Client.GetAll(context.TODO(), req)
	if err != nil {
		return err
	}
	data["posts"] = res.Posts
	return c.Render(http.StatusOK, "home.html", data)
}
