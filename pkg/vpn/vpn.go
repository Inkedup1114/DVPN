package vpn

import (
    "context"
)

type VPNServer interface {
    Start(ctx context.Context) error
    Stop() error
    AddClient(clientKey, clientIP string) error
    RemoveClient(clientKey string) error
    GetStats() *Stats
}

type VPNClient interface {
    Connect(serverKey, serverEndpoint string) error
    Disconnect() error
    IsConnected() bool
    GetStats() *Stats
}

type Stats struct {
    BytesSent     uint64
    BytesReceived uint64
    PacketsSent   uint64
    PacketsRecv   uint64
    Uptime        int64
}

type Config struct {
    Interface string
    Port      int
    PrivateKey string
    PublicKey  string
    ListenPort int
    MTU        int
}
