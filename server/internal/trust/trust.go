// Package trust implements a lightweight trust-score engine. It combines a
// client-supplied browser fingerprint with server-visible request signals to
// produce a 0-100 score used to annotate and gate brand generations.
package trust

import (
	"net"
	"net/http"
	"strings"
)

// Signals are the inputs to a trust evaluation for a single request.
type Signals struct {
	Fingerprint   string            // client-computed fingerprint hash
	Components    map[string]string // raw fingerprint components (screen, tz, ...)
	UserAgent     string
	Authenticated bool
	IP            string
}

// Score is the result of an evaluation.
type Score struct {
	Value   int      `json:"value"`   // 0-100, higher is more trustworthy
	Level   string   `json:"level"`   // "low" | "medium" | "high"
	Reasons []string `json:"reasons"` // human-readable contributing factors
}

// botMarkers are substrings that strongly indicate a non-browser or headless
// client.
var botMarkers = []string{
	"headlesschrome", "phantomjs", "selenium", "puppeteer", "playwright",
	"bot", "spider", "crawler", "curl/", "wget/", "python-requests",
	"go-http-client", "java/", "okhttp",
}

// Evaluate scores a request. The heuristics are intentionally transparent so
// the reasons can be surfaced to the user and tuned over time.
func Evaluate(s Signals) Score {
	value := 50
	reasons := []string{}

	// A rich, high-entropy fingerprint is the strongest positive signal.
	switch {
	case len(s.Fingerprint) >= 16:
		value += 20
		reasons = append(reasons, "stable browser fingerprint present")
	default:
		value -= 20
		reasons = append(reasons, "missing or weak browser fingerprint")
	}

	// More collected components => a real, fully-featured browser environment.
	if n := len(s.Components); n > 0 {
		bonus := n
		if bonus > 10 {
			bonus = 10
		}
		value += bonus
		reasons = append(reasons, "device exposes a full component set")
	}

	if s.Authenticated {
		value += 20
		reasons = append(reasons, "authenticated session")
	}

	ua := strings.ToLower(s.UserAgent)
	switch {
	case ua == "":
		value -= 15
		reasons = append(reasons, "no user-agent header")
	case containsAny(ua, botMarkers):
		value -= 30
		reasons = append(reasons, "automation/bot user-agent")
	case strings.Contains(ua, "mozilla"):
		value += 10
		reasons = append(reasons, "consumer browser user-agent")
	}

	// Requests from a private/loopback address are local dev, treated neutrally
	// but noted.
	if ip := net.ParseIP(s.IP); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		reasons = append(reasons, "request from local network")
	}

	value = clamp(value, 0, 100)
	return Score{Value: value, Level: level(value), Reasons: reasons}
}

// ClientIP extracts the best-effort client IP from a request, honoring a single
// hop of X-Forwarded-For.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func level(v int) string {
	switch {
	case v < 40:
		return "low"
	case v < 70:
		return "medium"
	default:
		return "high"
	}
}
