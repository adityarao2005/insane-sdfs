package punch

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	PeerID   string `json:"peerId,omitempty"`
	PeerAddr string `json:"peerAddr,omitempty"`
}

func encodeMessage(m Message) ([]byte, error) {
	return json.Marshal(m)
}

func decodeMessage(b []byte) (Message, error) {
	var m Message
	if err := json.Unmarshal(b, &m); err != nil {
		return Message{}, err
	}
	if m.Type == "" {
		return Message{}, fmt.Errorf("message type is required")
	}
	return m, nil
}
