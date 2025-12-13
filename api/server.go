package api

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/katana-project/attach"
)

type server struct {
	provider attach.Provider
}

func NewServer(provider attach.Provider) StrictServerInterface {
	return &server{provider}
}

func (*server) GetHealth(context.Context, GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "healthy", Timestamp: time.Now()}, nil
}

func (s *server) GetInstances(context.Context, GetInstancesRequestObject) (GetInstancesResponseObject, error) {
	var instances []Instance

	descs, err := s.provider.List()
	if err != nil {
		return nil, err
	}
	for _, desc := range descs {
		properties := make(map[string]string)

		vm, err := s.provider.Attach(desc)
		if err == nil {
			properties, err = vm.Properties()
			if err != nil {
				log.Error("failed to get properties", "id", desc.ID, "error", err)
			}
		} else {
			log.Error("failed to attach", "id", desc.ID, "error", err)
		}
		_ = vm.Close()

		instances = append(instances, Instance{
			Id:          desc.ID,
			DisplayName: desc.DisplayName,
			Properties:  properties,
		})
	}

	return GetInstances200JSONResponse(instances), nil
}
