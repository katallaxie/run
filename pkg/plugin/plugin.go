package plugin

import (
	"context"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	p "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/katallaxie/run/pkg/proto"
)

var enablePluginAutoMTLS = os.Getenv("RUN_DISABLE_PLUGIN_TLS") == ""

// Meta ...
type Meta struct {
	// Path ...
	Path string
	// Arguments ...
	Arguments []string
}

// ExecutableFile ...
func (m *Meta) ExecutableFile() (string, error) {
	// TODO: make this check for the executable file
	return m.Path, nil
}

func (m *Meta) Factory(ctx context.Context) Factory {
	return pluginFactory(ctx, m)
}

// GRPCTaskPlugin ...
type GRPCTaskPlugin struct {
	p.Plugin
	GRPCPlugin func() proto.PluginServer
}

// GRPCClient ...
func (p *GRPCTaskPlugin) GRPCClient(ctx context.Context, broker *p.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPlugin{
		client: proto.NewPluginClient(c),
		ctx:    ctx,
	}, nil
}

func (p *GRPCTaskPlugin) GRPCServer(broker *p.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPluginServer(s, p.GRPCPlugin())
	return nil
}

// GRPCPlugin ...
type GRPCPlugin struct {
	PluginClient *p.Client

	ctx    context.Context
	client proto.PluginClient
}

// Start ...
func (p *GRPCPlugin) Close() error {
	if p.PluginClient != nil {
		return nil
	}

	p.PluginClient.Kill()
	return nil
}

// Execute ...
func (p *GRPCPlugin) Execute(req ExecuteRequest) (ExecuteResponse, error) {
	r := new(proto.Execute_Request)
	r.Vars = req.Vars
	r.Args = req.Arguments

	_, err := p.client.Execute(p.ctx, r)
	if err != nil {
		return ExecuteResponse{}, err
	}

	return ExecuteResponse{}, nil
}

// Factory ...
type Factory func() (Plugin, error)

// Plugin ...
type Plugin interface {
	// Execute ...
	Execute(ExecuteRequest) (ExecuteResponse, error)
	// Close ...
	Close() error
}

// ExecuteRequest ...
type ExecuteRequest struct {
	Vars      map[string]string
	Arguments []string
}

// ExecuteResponse ...
type ExecuteResponse struct {
}

func pluginFactory(ctx context.Context, meta *Meta) Factory {
	return func() (Plugin, error) {
		f, err := meta.ExecutableFile()
		if err != nil {
			return nil, err
		}

		l := hclog.New(&hclog.LoggerOptions{
			Name:  meta.Path,
			Level: hclog.LevelFromString("DEBUG"),
		})

		cfg := &p.ClientConfig{
			Logger:           l,
			VersionedPlugins: VersionedPlugins,
			HandshakeConfig:  Handshake,
			AutoMTLS:         enablePluginAutoMTLS,
			Managed:          true,
			AllowedProtocols: []p.Protocol{p.ProtocolGRPC},
			Cmd:              exec.CommandContext(ctx, f, meta.Arguments...),
			SyncStderr:       l.StandardWriter(&hclog.StandardLoggerOptions{}),
			SyncStdout:       l.StandardWriter(&hclog.StandardLoggerOptions{}),
		}
		client := p.NewClient(cfg)

		rpc, err := client.Client()
		if err != nil {
			return nil, err
		}

		raw, err := rpc.Dispense(PluginName)
		if err != nil {
			return nil, err
		}

		p := raw.(*GRPCPlugin)
		p.PluginClient = client

		return p, nil
	}
}
