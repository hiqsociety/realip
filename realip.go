package realip

import (
	"errors"
	"net"
	"net/http"
	"strings"
)






//Copied from https://github.com/Ferluci/fasthttp-realip/


// Should use canonical format of the header key s
// https://golang.org/pkg/net/http/#CanonicalHeaderKey

// Header may return multiple IP addresses in the format: "client IP, proxy 1 IP, proxy 2 IP", so we take the the first one.
var xOriginalForwardedForHeader = http.CanonicalHeaderKey("X-Original-Forwarded-For")
var xForwardedForHeader = http.CanonicalHeaderKey("X-Forwarded-For")
var xForwardedHeader = http.CanonicalHeaderKey("X-Forwarded")
var forwardedForHeader = http.CanonicalHeaderKey("Forwarded-For")
var forwardedHeader = http.CanonicalHeaderKey("Forwarded")

// Standard headers used by Amazon EC2, Heroku, and others
var xClientIPHeader = http.CanonicalHeaderKey("X-Client-IP")

// Nginx proxy/FastCGI
var xRealIPHeader = http.CanonicalHeaderKey("X-Real-IP")

// Cloudflare.
// @see https://support.cloudflare.com/hc/en-us/articles/200170986-How-does-Cloudflare-handle-HTTP-Request-headers-
// CF-Connecting-IP - applied to every request to the origin.
var cfConnectingIPHeader = http.CanonicalHeaderKey("CF-Connecting-IP")

// Fastly CDN and Firebase hosting header when forwared to a cloud function
var fastlyClientIPHeader = http.CanonicalHeaderKey("Fastly-Client-Ip")

// Akamai and Cloudflare
var trueClientIPHeader = http.CanonicalHeaderKey("True-Client-Ip")















var cidrs []*net.IPNet

func init() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}
}

// isLocalAddress works by checking if the address is under private CIDR blocks.
// List of private CIDR blocks can be seen on :
//
// https://en.wikipedia.org/wiki/Private_network
//
// https://en.wikipedia.org/wiki/Link-local_address
func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

// FromRequest return client's real public IP address from http request headers.
func FromRequest(r *http.Request) string {
	xClientIP := r.Header.Get(xClientIPHeader)
	if len(xClientIP) != 0 {
		return string(xClientIP)
	}

	xOriginalForwardedFor := r.Header.Get(xOriginalForwardedForHeader)
	if xOriginalForwardedFor != nil {
		requestIP, err := retrieveForwardedIP(string(xOriginalForwardedFor))
		if err == nil {
			return requestIP
		}
	}

	xForwardedFor := r.Header.Get(xForwardedForHeader)
	if len(xForwardedFor) != 0 {
		requestIP, err := retrieveForwardedIP(string(xForwardedFor))
		if err == nil {
			return requestIP
		}
	}

	if ip, err := fromSpecialHeaders(r); err == nil {
		return ip
	}

	if ip, err := fromForwardedHeaders(r); err == nil {
		return ip
	}

	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	// Fetch header value
	xRealIP := r.Header.Get("X-Real-Ip")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	// If both empty, return IP from remote address
	if xRealIP == "" && xForwardedFor == "" {
		var remoteIP string

		// If there are colon in remote address, remove the port number
		// otherwise, return remote address as is
		if strings.ContainsRune(r.RemoteAddr, ':') {
			remoteIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		} else {
			remoteIP = r.RemoteAddr
		}

		return remoteIP
	}

	// Check list of IP in X-Forwarded-For and return the first global address
	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := isPrivateAddress(address)
		if !isPrivate && err == nil {
			return address
		}
	}

	// If nothing succeed, return X-Real-IP
	return xRealIP
}

// RealIP is depreciated, use FromRequest instead
func RealIP(r *http.Request) string {
	return FromRequest(r)
}




func fromSpecialHeaders(r *http.Request) (string, error) {
	ipHeaders := [...]string{cfConnectingIPHeader, fastlyClientIPHeader, trueClientIPHeader, xRealIPHeader}
	for _, iplHeader := range ipHeaders {
		if clientIP := r.Header.Get(iplHeader); len(clientIP) != 0 {
			return string(clientIP), nil
		}
	}
	return "", errors.New("can't get ip from special headers")
}

func fromForwardedHeaders(r *http.Request) (string, error) {
	forwardedHeaders := [...]string{xForwardedHeader, forwardedForHeader, forwardedHeader}
	for _, forwardedHeader := range forwardedHeaders {
		if forwarded := r.Header.Get(forwardedHeader); len(forwarded) != 0 {
			if clientIP, err := retrieveForwardedIP(string(forwarded)); err == nil {
				return clientIP, nil
			}
		}
	}
	return "", errors.New("can't get ip from forwarded headers")
}


func retrieveForwardedIP(forwardedHeader string) (string, error) {
	for _, address := range strings.Split(forwardedHeader, ",") {
		if len(address) > 0 {
			address = strings.TrimSpace(address)
			isPrivate, err := isPrivateAddress(address)
			switch {
			case !isPrivate && err == nil:
				return address, nil
			case isPrivate && err == nil:
				return "", errors.New("forwarded ip is private")
			default:
				return "", err
			}
		}
	}
	return "", errors.New("empty or invalid forwarded header")
}

