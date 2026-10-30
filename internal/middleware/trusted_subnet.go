package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware requires X-Real-IP to belong to the configured CIDR.
// The caller must only register it when a trusted subnet is configured.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	_, subnet, _ := net.ParseCIDR(trustedSubnet)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addresses := r.Header.Values("X-Real-IP")
			if len(addresses) != 1 || subnet == nil || !subnet.Contains(net.ParseIP(addresses[0])) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
