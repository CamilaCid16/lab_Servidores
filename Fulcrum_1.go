package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/CamilaCid16/lab" // Asegúrate de ajustar la importación según tu estructura de proyecto

	"google.golang.org/grpc"
)

// VectorClock management functions
type FulcrumServer struct {
	pb.UnimplementedFulcrumServer
	ServerId     int
	VectorClocks map[string][]int32
	Logs         map[string][]LogEntry
	mutex        sync.Mutex
}

// Log entry structure
type LogEntry struct {
	Action string
	Sector string
	Base   string
	Value  int32
}

// NewFulcrumServer creates a new instance of FulcrumServer
func NewFulcrumServer(serverId int) *FulcrumServer {
	return &FulcrumServer{
		ServerId:     serverId,
		VectorClocks: make(map[string][]int32),
		Logs:         make(map[string][]LogEntry),
	}
}

// UpdateVectorClock updates the vector clock for the given sector
func (s *FulcrumServer) UpdateVectorClock(sector string) {
	if _, ok := s.VectorClocks[sector]; !ok {
		s.VectorClocks[sector] = make([]int32, 3) // Assuming 3 servers
	}
	s.VectorClocks[sector][s.ServerId]++
}

// UpdateFile writes content to a file named after the sector
func UpdateFile(sector, content string) {
	filename := fmt.Sprintf("%s.txt", sector)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
}

// AddBase implements the AddBase RPC method
func (s *FulcrumServer) AddBase(ctx context.Context, req *pb.BaseRequest) (*pb.BaseResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sector := req.GetSector()
	base := req.GetBase()
	value := req.GetValue()

	UpdateFile(sector, fmt.Sprintf("%s %d\n", base, value))
	s.UpdateVectorClock(sector)
	s.Logs[sector] = append(s.Logs[sector], LogEntry{Action: "AgregarBase", Sector: sector, Base: base, Value: value})

	return &pb.BaseResponse{Success: true, VectorClock: s.VectorClocks[sector]}, nil
}

// Implement other RPC methods similarly

// StartServer starts a gRPC server on the given port
func StartServer(port string, serverId int) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFulcrumServer(s, NewFulcrumServer(serverId))
	log.Printf("Server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	port := ":50051"
	serverId := 0
	log.Printf("Starting server 1 on port %s", port)
	StartServer(port, serverId)
}
