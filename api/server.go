package api

import (
	"context"
	"time"
)

type server struct {
}

func NewServer() StrictServerInterface {
	return &server{}
}

func (*server) GetHealth(context.Context, GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "healthy", Timestamp: time.Now()}, nil
}
