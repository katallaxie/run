package main

import (
	"context"
	"log"

	"github.com/katallaxie/run/pkg/plugin"
	"github.com/katallaxie/run/pkg/proto"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCPluginFunc: func() proto.PluginServer {
			return &server{}
		},
	})
}

type server struct {
	proto.UnimplementedPluginServer
}

// Start ...
func (s *server) Execute(ctx context.Context, req *proto.Execute_Request) (*proto.Execute_Response, error) {
	log.Print(req)

	return &proto.Execute_Response{}, nil
}

// Stop ...
func (s *server) Stop(ctx context.Context, req *proto.Stop_Request) (*proto.Stop_Response, error) {
	log.Print(req)

	return &proto.Stop_Response{}, nil
}
