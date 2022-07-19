package main

import (
	"context"
	"fmt"

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
	resp := &proto.Execute_Response{}

	p := NewGit()
	_ = NewArchive()

	err := p.CloneWithContext(ctx, req.Vars["url"], req.Vars["folder"])
	if err != nil {
		resp.Diagnostic = []*proto.Diagnostic{
			proto.DiagnosticFromError(err),
		}
		return resp, err
	}

	return &proto.Execute_Response{}, nil
}

// Stop ...
func (s *server) Stop(ctx context.Context, req *proto.Stop_Request) (*proto.Stop_Response, error) {
	fmt.Println("Stop")

	return &proto.Stop_Response{}, nil
}
