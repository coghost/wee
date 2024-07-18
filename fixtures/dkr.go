package fixtures

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	_image         = `mitmproxy/mitmproxy`
	_containerPort = "8080/tcp"
)

type Container struct {
	URI     string
	volumes []string

	ctx    context.Context
	client *client.Client
	id     string
}

// NewContainer creates a new mongodb container, please remember to call `.Clear()` to remove it.
//
//	@param ctx
//	@param args `container name, use the first arg passed in as container name`
//	@return *Container
func NewContainer(ctx context.Context, name string, args ...string) *Container {
	dkr, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := dkr.ContainerCreate(
		ctx,
		&container.Config{
			Image: _image,
			ExposedPorts: nat.PortSet{
				_containerPort: {},
			},
			Cmd: args,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				_containerPort: []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "0",
					},
				},
			},
		},
		nil,
		nil,
		name,
	)
	if err != nil {
		panic(err)
	}

	containerID := resp.ID

	err = dkr.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		panic(err)
	}

	res, err := dkr.ContainerInspect(ctx, containerID)
	if err != nil {
		panic(err)
	}

	vols := []string{}
	for _, m := range res.Mounts {
		vols = append(vols, m.Name)
	}

	pb := res.NetworkSettings.Ports[_containerPort][0]

	cont := &Container{
		client:  dkr,
		ctx:     ctx,
		id:      resp.ID,
		URI:     fmt.Sprintf("%s:%s", pb.HostIP, pb.HostPort),
		volumes: vols,
	}

	return cont
}

// Clear removes the container and volumes created
//
//	@receiver m
//	@param ctx
func (m *Container) Clear() {
	if m.id == "" {
		return
	}

	err := m.client.ContainerRemove(
		m.ctx,
		m.id,
		container.RemoveOptions{
			Force: true,
		})
	if err != nil {
		panic(err)
	}

	for _, vol := range m.volumes {
		err = m.client.VolumeRemove(m.ctx, vol, true)
		if err != nil {
			panic(err)
		}
	}
}
