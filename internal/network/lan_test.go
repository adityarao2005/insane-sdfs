package network

import (
	"net"
	"testing"
)

func TestIsPrivateLANIP(t *testing.T) {
	cases := []struct {
		IP   string
		Want bool
	}{
		{IP: "127.0.0.1", Want: true},
		{IP: "10.1.2.3", Want: true},
		{IP: "172.16.8.1", Want: true},
		{IP: "192.168.1.8", Want: true},
		{IP: "8.8.8.8", Want: false},
		{IP: "1.1.1.1", Want: false},
	}
	for _, tc := range cases {
		ip := net.ParseIP(tc.IP)
		if got := IsPrivateLANIP(ip); got != tc.Want {
			t.Fatalf("ip=%s got=%v want=%v", tc.IP, got, tc.Want)
		}
	}
}
