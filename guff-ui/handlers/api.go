package handlers

import (
	"context"
	"encoding/json"
	"guff-ui/pb"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetAll : get All posts
func (a *API) GetAll(c echo.Context) error {
	cachedData, err := a.Cache.Get("getAll")
	if err == nil {
		m := new(map[string]interface{})
		err = json.Unmarshal(cachedData, m)
		if err == nil {
			return c.JSON(http.StatusOK, m)
		}
	}
	req := &pb.GetAllRequest{}
	res, err := a.Client.GetAll(context.TODO(), req)
	if err != nil {
		return err
	}
	rm := echo.Map{"data": res.Posts, "meta": res.PostMeta}
	b, err := json.Marshal(rm)
	if err == nil {
		a.Cache.Set("getAll", b)
	}
	return c.JSON(http.StatusOK, rm)
}
