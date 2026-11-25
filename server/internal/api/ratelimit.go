package api

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type clientEntry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// uploadLimiter enforces a per-client-IP token bucket on snapshot uploads and
// evicts idle clients so memory stays bounded.
type uploadLimiter struct {
	mu            sync.Mutex
	limit         rate.Limit
	burst         int
	trusted       []netip.Prefix
	now           func() time.Time
	idleTTL       time.Duration
	sweepInterval time.Duration
	lastSweep     time.Time
	clients       map[string]*clientEntry
}

func newUploadLimiter(perHour, burst int, trusted []netip.Prefix, now func() time.Time) *uploadLimiter {
	if perHour <= 0 {
		return nil
	}
	if burst < 1 {
		burst = 1
	}
	if now == nil {
		now = time.Now
	}
	lim := rate.Limit(float64(perHour) / time.Hour.Seconds())
	// Keep an entry at least as long as it takes an empty bucket to refill
	// completely, so evicting an idle client never loses rate-limit state.
	refill := time.Duration(float64(burst) / float64(lim) * float64(time.Second))
	idleTTL := max(refill, 10*time.Minute)
	sweepInterval := min(idleTTL, 5*time.Minute)

	return &uploadLimiter{
		limit:         lim,
		burst:         burst,
		trusted:       trusted,
		now:           now,
		idleTTL:       idleTTL,
		sweepInterval: sweepInterval,
		lastSweep:     now(),
		clients:       make(map[string]*clientEntry),
	}
}

// allow checks whether the client behind r may upload right now. When denied,
// retryAfter is the wait until the next token is available.
func (l *uploadLimiter) allow(r *http.Request) (bool, time.Duration) {
	key := l.clientKey(r)
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) >= l.sweepInterval {
		for k, e := range l.clients {
			if now.Sub(e.lastSeen) >= l.idleTTL {
				delete(l.clients, k)
			}
		}
		l.lastSweep = now
	}

	e, ok := l.clients[key]
	if !ok {
		e = &clientEntry{lim: rate.NewLimiter(l.limit, l.burst)}
		l.clients[key] = e
	}
	e.lastSeen = now

	res := e.lim.ReserveN(now, 1)
	delay := res.DelayFrom(now)
	if delay > 0 {
		res.CancelAt(now)
		return false, delay
	}
	return true, 0
}

func (l *uploadLimiter) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.clients)
}

func (l *uploadLimiter) clientKey(r *http.Request) string {
	remoteIP, ok := parseRemoteAddr(r.RemoteAddr)
	if !ok {
		return r.RemoteAddr
	}
	clientIP := remoteIP
	if l.isTrusted(remoteIP) {
		if xffIP, ok := l.xffClientIP(r.Header.Values("X-Forwarded-For")); ok {
			clientIP = xffIP
		}
	}
	return ipKey(clientIP)
}

func (l *uploadLimiter) isTrusted(ip netip.Addr) bool {
	for _, p := range l.trusted {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// xffClientIP returns the right-most X-Forwarded-For address that is not a
// trusted proxy.
func (l *uploadLimiter) xffClientIP(headers []string) (netip.Addr, bool) {
	var parts []string
	for _, h := range headers {
		parts = append(parts, strings.Split(h, ",")...)
	}
	for i := len(parts) - 1; i >= 0; i-- {
		s := strings.TrimSpace(parts[i])
		if s == "" {
			continue
		}
		ip, err := netip.ParseAddr(s)
		if err != nil {
			return netip.Addr{}, false
		}
		ip = ip.Unmap()
		if !l.isTrusted(ip) {
			return ip, true
		}
	}
	return netip.Addr{}, false
}

func parseRemoteAddr(addr string) (netip.Addr, bool) {
	if ap, err := netip.ParseAddrPort(addr); err == nil {
		return ap.Addr().Unmap(), true
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		if ip, err := netip.ParseAddr(host); err == nil {
			return ip.Unmap(), true
		}
	}
	if ip, err := netip.ParseAddr(addr); err == nil {
		return ip.Unmap(), true
	}
	return netip.Addr{}, false
}

func ipKey(ip netip.Addr) string {
	ip = ip.Unmap()
	if ip.Is6() {
		return netip.PrefixFrom(ip, 64).Masked().String()
	}
	return ip.String()
}

func formatRetryAfter(d time.Duration) (int, string) {
	secs := max(1, int(math.Ceil(d.Seconds())))
	switch {
	case secs < 60:
		return secs, fmt.Sprintf("%ds", secs)
	case secs%60 == 0:
		return secs, fmt.Sprintf("%dm", secs/60)
	default:
		return secs, fmt.Sprintf("%dm%ds", secs/60, secs%60)
	}
}
