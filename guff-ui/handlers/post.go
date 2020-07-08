package handlers

import (
	"bufio"
	"context"
	"guff-ui/pb"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreatePost :
func (w *Web) CreatePost(c echo.Context) error {
	data := echo.Map{"csrf_token": c.Get("csrf")}
	if c.Request().Method == "GET" {
		return c.Render(200, "create.html", data)
	}

	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["photos"]
	var photos []*pb.Photo
	for _, file := range files {
		// Source
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		stream, err := w.Client.UploadPhoto(context.TODO())
		if err != nil {
			return err
		}
		uploadMeta := &pb.UploadPhotoRequest_PhotoMeta{}
		uploadMetaReq := &pb.UploadPhotoRequest{Data: uploadMeta}
		err = stream.Send(uploadMetaReq)
		if err != nil {
			return err
		}
		reader := bufio.NewReader(src)
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			data := &pb.UploadPhotoRequest_ChunkData{ChunkData: buffer[:n]}
			uploadPhotoReq := &pb.UploadPhotoRequest{Data: data}
			err = stream.Send(uploadPhotoReq)
			if err != nil {
				return err
			}
		}
		uploadRes, err := stream.CloseAndRecv()
		if err != nil {
			return err
		}
		photos = append(photos, &pb.Photo{Name: uploadRes.FileName})
	}
	//photo upload complete
	//create post
	post := &pb.Post{}
	post.Title = c.FormValue("title")
	post.Content = c.FormValue("description")
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil {
		//todo show user's fault
		return err
	}
	post.Price = price
	status := c.FormValue("availability")
	switch status {
	case "0":
		post.Status = pb.Status_IN_STOCK
		break
	case "1":
		post.Status = pb.Status_PRE_ORDER
		break
	case "2":
		post.Status = pb.Status_OUT_OF_STOCK
		break
	default:
		post.Status = pb.Status_IN_STOCK
		break
	}
	createReq := &pb.CreateRequest{Photos: photos, Post: post}
	_, err = w.Client.Create(context.Background(), createReq)
	if err != nil {
		return err
	}
	w.Cache.Delete("getAll")
	data["msg"] = "Successfully Created"
	return c.Render(http.StatusCreated, "create.html", data)
}
