package post

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"grng.dev/guff/pb"
	"grng.dev/guff/utils"
)

// Service : post
type Service struct {
	DB      *mongo.Database
	Storage utils.Storage
}

// GetAll : Get all post
func (s *Service) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	dbCtx := context.TODO()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"createdon": -1})
	cur, err := s.DB.Collection("posts").Find(dbCtx, bson.M{}, findOptions)
	if err != nil {
		log.Errorf("mongodb error: %s", err)
		return nil, err
	}
	res := &pb.GetAllResponse{}
	meta := &pb.PostMeta{}
	meta.CurrentPage = req.GetPage()
	meta.TotalPages = 1
	posts := []*pb.Post{}
	defer cur.Close(dbCtx)
	for cur.Next(dbCtx) {
		res := new(pb.Post)
		err := cur.Decode(res)
		if err != nil {
			log.Errorf("failed to decode mongo result: %s", err)
			return nil, err
		}
		photoCtx := context.TODO()
		photoCur, err := s.DB.Collection("photos").Find(photoCtx, bson.M{"post": res.Uuid})
		if err != nil {
			return nil, err
		}
		defer photoCur.Close(photoCtx)
		for photoCur.Next(photoCtx) {
			p := new(pb.Photo)
			err = photoCur.Decode(p)
			if err != nil {
				return nil, err
			}
			url, err := s.Storage.SignedURL(p.Name)
			if err != nil {
				return nil, err
			}
			res.Photos = append(res.Photos, url)
		}
		posts = append(posts, res)
	}
	res.Posts = posts
	res.PostMeta = meta
	return res, nil
}

// Create : creates new post
func (s *Service) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	postCollecton := s.DB.Collection("posts")
	ses, err := s.DB.Client().StartSession()
	if err != nil {
		log.Error("failed to start db session : %s", err)
		return nil, err
	}
	if err = ses.StartTransaction(); err != nil {
		log.Error("failed to start transaction : %s", err)
		return nil, err
	}
	postUUID := uuid.New().String()
	sesContext := context.TODO()
	err = mongo.WithSession(sesContext, ses, func(sc mongo.SessionContext) error {
		req.Post.Uuid = postUUID
		req.Post.CreatedOn = time.Now().String()
		req.Post.UpdatedOn = time.Now().String()
		_, err := postCollecton.InsertOne(sc, req.Post)
		if err != nil {
			ses.AbortTransaction(sc)
			return err
		}
		var photos []interface{}
		for _, p := range req.Photos {
			p.Uuid = uuid.New().String()
			p.Post = postUUID
			photos = append(photos, p)
		}
		photoCollection := s.DB.Collection("photos")
		_, err = photoCollection.InsertMany(sc, photos)
		if err != nil {
			log.Error("Aborting")
			ses.AbortTransaction(sc)
			return err
		}
		if err = ses.CommitTransaction(sc); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("failed to commit transaction : %s", err)
		return &pb.CreateResponse{Success: false}, err
	}
	ses.EndSession(sesContext)
	return &pb.CreateResponse{Success: true, Id: postUUID}, nil
}

// UploadPhoto : uploads photo to gcp bucket
func (s *Service) UploadPhoto(stream pb.PostService_UploadPhotoServer) error {
	meta, err := stream.Recv()
	if err != nil {
		return err
	}
	photoData := bytes.Buffer{}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := req.GetChunkData()
		// Todo : max size check
		_, err = photoData.Write(chunk)
		if err != nil {
			return err
		}
	}
	objectName := uuid.New().String()
	objectName += "_" + meta.GetPhotoMeta().GetFileName()
	err = s.Storage.Store(objectName, photoData.Bytes(), utils.ContentOptions{ContentType: meta.GetPhotoMeta().ContentType})
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	res := &pb.UploadPhotoResponse{Success: true, FileName: objectName}
	return stream.SendAndClose(res)
}

// Delete : delete the post
func (s *Service) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	dbContext := context.TODO()
	ses, err := s.DB.Client().StartSession()
	if err != nil {
		log.Error("failed to start db session : %s", err)
		return nil, err
	}
	if err = ses.StartTransaction(); err != nil {
		log.Error("failed to start transaction : %s", err)
		return nil, err
	}
	err = mongo.WithSession(dbContext, ses, func(sc mongo.SessionContext) error {
		_, err = s.DB.Collection("posts").DeleteOne(sc, bson.M{"uuid": req.Uuid})
		if err != nil {
			ses.AbortTransaction(sc)
			return err
		}
		_, err = s.DB.Collection("photos").DeleteMany(sc, bson.M{"post": req.Uuid})
		if err != nil {
			ses.AbortTransaction(sc)
			return err
		}
		if err = ses.CommitTransaction(sc); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error("failed to commit transaction : %s", err)
		return &pb.DeleteResponse{Success: true}, err
	}
	ses.EndSession(dbContext)
	return &pb.DeleteResponse{Success: true}, nil
}
