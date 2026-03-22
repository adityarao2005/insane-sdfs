package network

import "net"

// IsPrivateLANIP returns true for addresses that should be treated as local/LAN scoped.
func IsPrivateLANIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return true
	}
	if privateRange10.Contains(ip) || privateRange172.Contains(ip) || privateRange192.Contains(ip) {
		return true
	}
	if uniqueLocalIPv6.Contains(ip) {
		return true
	}
	return false
}

var (
	privateRange10  = mustCIDR("10.0.0.0/8")
	privateRange172 = mustCIDR("172.16.0.0/12")
	privateRange192 = mustCIDR("192.168.0.0/16")
	uniqueLocalIPv6 = mustCIDR("fc00::/7")
)

func mustCIDR(c string) *net.IPNet {
	_, n, err := net.ParseCIDR(c)
	if err != nil {
		panic(err)
	}
	return n
}
