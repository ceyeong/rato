package utils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// Storage :
type Storage interface {
	// Store : stores the provided data
	Store(name string, data []byte, options ContentOptions) error
	// SignedURL : generate signed url
	SignedURL(name string) (string, error)
}

// ContentOptions :
type ContentOptions struct {
	ContentType string
}

type gcpStorage struct {
	credentials map[string]string
}

func (s *gcpStorage) initCredential() error {
	keyFile, err := os.Open(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		return err
	}
	defer keyFile.Close()
	byteVals, err := ioutil.ReadAll(keyFile)
	if err != nil {
		return err
	}
	var result map[string]string
	err = json.Unmarshal(byteVals, &result)
	if err != nil {
		return err
	}
	s.credentials = result
	return nil
}
func (s *gcpStorage) SignedURL(fileName string) (string, error) {
	if s.credentials == nil {
		err := s.initCredential()
		if err != nil {
			return "", err
		}
	}
	return storage.SignedURL(os.Getenv("STORAGE_BUCKET"), fileName, &storage.SignedURLOptions{
		GoogleAccessID: s.credentials["client_email"],
		PrivateKey:     []byte(s.credentials["private_key"]),
		Method:         "GET",
		Expires:        time.Now().Add(time.Hour * 24 * 30),
	})
}

func (s *gcpStorage) Store(name string, data []byte, options ContentOptions) error {
	storageContext := context.Background()
	storageClient, err := storage.NewClient(storageContext)
	if err != nil {
		return err
	}
	wc := storageClient.Bucket(os.Getenv("STORAGE_BUCKET")).Object(name).NewWriter(context.Background())
	if options.ContentType != "" {
		wc.ContentType = options.ContentType
	}
	wc.Write(data)
	err = wc.Close()
	if err != nil {
		return err
	}
	return nil
}

// NewGCPStorage :
func NewGCPStorage() Storage {
	return new(gcpStorage)
}
