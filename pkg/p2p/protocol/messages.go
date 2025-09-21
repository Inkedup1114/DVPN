package protocol

import (
    "encoding/json"
    "time"
)

type MessageType uint8

const (
    MsgPing MessageType = iota
    MsgPong
    MsgNodeAdvertisement
    MsgNodeListRequest
    MsgNodeListResponse
    MsgBandwidthProof
    MsgMicropaymentUpdate
    MsgMicropaymentClose
    MsgBlockAnnouncement
    MsgTransactionAnnouncement
)

type Message struct {
    Type      MessageType     `json:"type"`
    Timestamp int64          `json:"timestamp"`
    Data      json.RawMessage `json:"data"`
    Signature []byte         `json:"signature"`
}

type PingMessage struct {
    Timestamp int64 `json:"timestamp"`
    Nonce     uint64 `json:"nonce"`
}

type PongMessage struct {
    Timestamp int64 `json:"timestamp"`
    Nonce     uint64 `json:"nonce"`
}

type NodeAdvertisementMessage struct {
    NodeInfo struct {
        PubKey             []byte `json:"pubkey"`
        IPAddress          string `json:"ip_address"`
        Port               uint16 `json:"port"`
        LocationHash       string `json:"location_hash"`
        AvailableBandwidth uint64 `json:"available_bandwidth_mbps"`
        MinPricePerMB      uint64 `json:"min_price_per_mb"`
        NodeType           string `json:"node_type"`
    } `json:"node_info"`
    ExpiryBlock uint64 `json:"expiry_block"`
}

type NodeListRequestMessage struct {
    LocationFilter string `json:"location_filter,omitempty"`
    MinBandwidth   uint64 `json:"min_bandwidth,omitempty"`
    MaxPrice       uint64 `json:"max_price,omitempty"`
    Limit          uint32 `json:"limit"`
}

type NodeListResponseMessage struct {
    Nodes []NodeAdvertisementMessage `json:"nodes"`
}

type BandwidthProofMessage struct {
    SessionIDHash []byte                 `json:"session_id_hash"`
    BytesSent     uint64                 `json:"bytes_sent"`
    BytesReceived uint64                 `json:"bytes_received"`
    AuditSigs     []AuditSignature       `json:"audit_sigs"`
}

type AuditSignature struct {
    AuditorPubKey []byte `json:"auditor_pubkey"`
    Signature     []byte `json:"signature"`
    Timestamp     int64  `json:"timestamp"`
}

type MicropaymentUpdateMessage struct {
    ChannelID    []byte `json:"channel_id"`
    SequenceNum  uint64 `json:"sequence_num"`
    Amount       uint64 `json:"amount"`
    RecipientSig []byte `json:"recipient_sig"`
    SenderSig    []byte `json:"sender_sig"`
}

type MicropaymentCloseMessage struct {
    ChannelID   []byte `json:"channel_id"`
    FinalAmount uint64 `json:"final_amount"`
    Signatures  struct {
        Sender    []byte `json:"sender"`
        Recipient []byte `json:"recipient"`
    } `json:"signatures"`
}

func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
    dataBytes, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    
    return &Message{
        Type:      msgType,
        Timestamp: time.Now().Unix(),
        Data:      dataBytes,
    }, nil
}

func (m *Message) DecodeData(dest interface{}) error {
    return json.Unmarshal(m.Data, dest)
}
