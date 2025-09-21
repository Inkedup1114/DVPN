package wireguard

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WgClient abstracts the wgctrl.Client used by the server so it can be mocked.
type WgClient interface {
	ConfigureDevice(iface string, cfg wgtypes.Config) error
	Device(iface string) (*wgtypes.Device, error)
	Close() error
}

// DefaultWgClient wraps the real wgctrl.Client
type DefaultWgClient struct {
	c *wgctrl.Client
}

func NewDefaultWgClient() (*DefaultWgClient, error) {
	c, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &DefaultWgClient{c: c}, nil
}

func (d *DefaultWgClient) ConfigureDevice(iface string, cfg wgtypes.Config) error {
	return d.c.ConfigureDevice(iface, cfg)
}

func (d *DefaultWgClient) Device(iface string) (*wgtypes.Device, error) {
	return d.c.Device(iface)
}

func (d *DefaultWgClient) Close() error {
	return d.c.Close()
}
