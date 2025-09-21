package wireguard

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// fakeRunner records executed commands for tests
type fakeRunner struct {
	calls [][]string
}

func (f *fakeRunner) Run(name string, args ...string) error {
	call := append([]string{name}, args...)
	f.calls = append(f.calls, call)
	return nil
}

// recordingFakeWgClient records the last ConfigureDevice call and counts calls
type recordingFakeWgClient struct {
	lastCfg wgtypes.Config
	calls   int
}

func (f *recordingFakeWgClient) ConfigureDevice(iface string, cfg wgtypes.Config) error {
	f.lastCfg = cfg
	f.calls++
	return nil
}

func (f *recordingFakeWgClient) Device(iface string) (*wgtypes.Device, error) {
	return &wgtypes.Device{Peers: []wgtypes.Peer{}}, nil
}

func (f *recordingFakeWgClient) Close() error { return nil }

// fakeWgClient used by some tests
type fakeWgClient struct {
	configured bool
	device     *wgtypes.Device
}

func (f *fakeWgClient) ConfigureDevice(iface string, cfg wgtypes.Config) error {
	f.configured = true
	return nil
}

func (f *fakeWgClient) Device(iface string) (*wgtypes.Device, error) {
	if f.device == nil {
		return &wgtypes.Device{Peers: []wgtypes.Peer{}}, nil
	}
	return f.device, nil
}

func (f *fakeWgClient) Close() error { return nil }
