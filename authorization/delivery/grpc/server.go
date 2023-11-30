package delivery_auth_grpc

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/repository/profile"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/repository/session"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	pb "github.com/go-park-mail-ru/2023_2_Vkladyshi/authorization/proto"
)

type Config struct {
	Port           string `yaml:"port"`
	ConnectionType string `yaml:"connection_type"`
}

type server struct {
	pb.UnimplementedAuthorizationServer
	userRepo    *profile.RepoPostgre
	sessionRepo *session.SessionRepo
	lg          *slog.Logger
}

func (s *server) GetId(ctx context.Context, req *pb.FindIdRequest) (*pb.FindIdResponse, error) {
	login, err := s.sessionRepo.GetUserLogin(ctx, req.Sid, s.lg)
	if err != nil {
        return nil, err
    }
	
    id, err := s.userRepo.GetUserProfileId(login)
    if err != nil {
        return nil, err
    }
    return &pb.FindIdResponse{
        Value: id,
    }, nil
}

func (s *server) GetIdsAndPaths(ctx context.Context, req *pb.IdsAndPathsListRequest) (*pb.IdsAndPathsResponse, error) {
    ids, paths, err := s.userRepo.GetIdsAndPaths()
    if err != nil {
        return nil, err
    }
    return &pb.IdsAndPathsResponse{
        Ids:   ids,
        Paths: paths,
    }, nil
}


func (s *server) GetAuthorizationStatus(ctx context.Context, req *pb.AuthorizationCheckRequest) (*pb.AuthorizationCheckResponse, error) {
    status, err := s.sessionRepo.CheckActiveSession(ctx, req.Sid, s.lg)
    if err != nil {
        return nil, err
    }
    return &pb.AuthorizationCheckResponse{
        Status: status,
    }, nil
}


func ListenAndServeGrpc(l *slog.Logger) {
	filename := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	file, err := os.Open(*filename)
	if err != nil {
		l.Error("failed to open config file: %v", err)
		return
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		l.Error("failed to parse config file: %v", err)
		return
	}

	lis, err := net.Listen(config.ConnectionType, ":"+config.Port)
	if err != nil {
		l.Error("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterAuthorizationServer(s, &server{
		lg: l,
	})
	if err := s.Serve(lis); err != nil {
		l.Error("failed to serve: %v", err)
		return
	}
}
