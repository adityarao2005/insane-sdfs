package punch

import "testing"

func TestEncodeDecodeMessage(t *testing.T) {
	b, err := encodeMessage(Message{Type: "register", Token: "abc", PeerID: "peer-1"})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	msg, err := decodeMessage(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg.Type != "register" || msg.Token != "abc" || msg.PeerID != "peer-1" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestDecodeRequiresType(t *testing.T) {
	_, err := decodeMessage([]byte(`{"token":"x"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
