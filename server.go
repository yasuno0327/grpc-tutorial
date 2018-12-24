package main

import (
	"context"
	pb "grpc-tutorial/routeguide"
	"math"
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

func (s *routeGuideServer) ListFeature(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}
	return nil
}

// 範囲内にあるかどうか
func inRange(point *pb.Point, rect *pb.Rectangle) bool {
	// left: 経度の最小値 right: 経度の最大値 top: 緯度の最大値 bottom: 緯度の最小値
	left := math.Min(float64(rect.Lo.Longititude), float64(rect.Hi.Longititude))
	right := math.Max(float64(rect.Lo.Longititude), float64(rect.Hi.Longititude))
	top := math.Max(float64(rect.Lo.Lattitude), float64(rect.Hi.Lattitude))
	bottom := math.Min(float64(rect.Lo.Lattitude), float64(rect.Hi.Lattitude))

	if float64(point.Longititude) >= left &&
		float64(point.Longititude) <= right &&
		float64(point.Lattitude) <= top &&
		float64(point.Lattitude) >= bottom {
		return true
	}
	return false
}
