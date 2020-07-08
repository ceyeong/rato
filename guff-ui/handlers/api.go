package handlers

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"guff-ui/pb"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetAll : get All posts
func (a *API) GetAll(c echo.Context) error {
	cachedData, err := a.Cache.Get("getAll")
	var cache interface{}
	if err == nil {
		if err = json.Unmarshal(cachedData, &cache); err != nil {
			return c.JSON(http.StatusOK, cache)
		}
	}
	req := &pb.GetAllRequest{}
	res, err := a.Client.GetAll(context.TODO(), req)
	if err != nil {
		return err
	}
	rm := echo.Map{"data": res.Posts, "meta": res.PostMeta}
	b, err := getByte(rm)
	if err == nil {
		a.Cache.Set("getAll", b)
	}
	return c.JSON(http.StatusOK, rm)
}
func getByte(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
