package middleware

import "net/http"

// SecurityHeaders applies security headers to all HTTP responses.
// These headers protect against common web vulnerabilities including
// clickjacking, MIME sniffing, and content injection.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' https://avatars.githubusercontent.com https://lh3.googleusercontent.com data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
