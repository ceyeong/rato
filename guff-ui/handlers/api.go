package handlers

import (
	"context"
	"guff-ui/pb"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetAll : get All posts
func (a *API) GetAll(c echo.Context) error {
	req := &pb.GetAllRequest{}
	res, err := a.Client.GetAll(context.TODO(), req)
	if err != nil {
		return err
	}
	rm := echo.Map{"data": res.Posts, "meta": res.PostMeta}
	return c.JSON(http.StatusOK, rm)
}
