package main

import (
	"context"
	pb "grpc-tutorial/routeguide"
	"sync"

	"github.com/golang/protobuf/proto"
)

func main() {

}

type routeGuideServer struct {
	savedFeatures []*pb.Feature

	mu         sync.Mutex
	routeNotes map[string][]*pb.RouteNote
}

func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (*pb.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}

	// Featureが見つからない場合は空のFeatureを返す
	return &pb.Feature{Location: point}, nil
}
