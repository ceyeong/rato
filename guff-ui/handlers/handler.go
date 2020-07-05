package handlers

import (
	"guff-ui/pb"
)

// Web : web ui handler
type Web struct {
	Client pb.PostServiceClient
}

// API : handler for api
type API struct {
	Client pb.PostServiceClient
}
