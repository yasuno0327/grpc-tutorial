package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pb "grpc-tutorial/routeguide"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "testdata/route_guide_db.json", "A json file containing a list of features")
	port       = flag.Int("port", 8080, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{routeNotes: make(map[string][]*pb.RouteNote)}
	s.loadFeatures(*jsonDBFile)
	return s
}

// loadFeatures loads features from a JSON file.
func (s *routeGuideServer) loadFeatures(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
	if err := json.Unmarshal(file, &s.savedFeatures); err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
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

func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
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

func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()
	for {
		point, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}
		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}
}

func calcDistance(p1 *pb.Point, p2 *pb.Point) int32 {
	const CordFactor float64 = 1e7
	const R float64 = float64(6371000)
	lat1 := toRadians(float64(p1.Lattitude) / CordFactor)
	lat2 := toRadians(float64(p2.Lattitude) / CordFactor)
	lng1 := toRadians(float64(p1.Longititude) / CordFactor)
	lng2 := toRadians(float64(p2.Longititude) / CordFactor)
	dlat := lat2 - lat1
	dlng := lng2 - lng1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}

func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := serialize(in.Location)

		s.mu.Lock()
		s.routeNotes[key] = append(s.routeNotes[key], in)
		rn := make([]*pb.RouteNote, len(s.routeNotes[key]))
		copy(rn, s.routeNotes[key])
		s.mu.Unlock()
		for _, note := range rn {
			if err := stream.Send(note); err != nil {
				return err
			}
		}
	}
}

func serialize(point *pb.Point) string {
	return fmt.Sprintf("%d %d", point.Lattitude, point.Longititude)
}
