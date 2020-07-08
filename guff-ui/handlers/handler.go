package handlers

import (
	"guff-ui/pb"

	"github.com/allegro/bigcache"
)

// Web : web ui handler
type Web struct {
	Client pb.PostServiceClient
	Cache  *bigcache.BigCache
}

// API : handler for api
type API struct {
	Client pb.PostServiceClient
	Cache  *bigcache.BigCache
}
