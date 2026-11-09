package httpserver

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const (
	// maxTargets is the maximum number of targets accepted in one scrape.
	maxTargets = 100
	// maxTargetLength is the maximum length of a single target parameter.
	maxTargetLength = 260
	defaultTLSPort  = "443"
)

var hostnameRe = regexp.MustCompile(`^[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?(\.[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?)*\.?$`)

// parseTarget validates a user supplied target ("host", "host:port",
// "[v6]:port") and returns the address to dial. Anything that is not a plain
// host and port (paths, credentials, queries, control characters...) is
// rejected, so a target can never be turned into an arbitrary request.
func parseTarget(target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("empty target")
	}
	if len(target) > maxTargetLength {
		return "", fmt.Errorf("target too long")
	}

	host, port := target, defaultTLSPort
	if h, p, err := net.SplitHostPort(target); err == nil {
		host, port = h, p
	} else if strings.HasPrefix(target, "[") && strings.HasSuffix(target, "]") {
		host = target[1 : len(target)-1]
	}

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNum == 0 {
		return "", fmt.Errorf("invalid port")
	}

	if net.ParseIP(host) == nil && !hostnameRe.MatchString(host) {
		return "", fmt.Errorf("invalid host")
	}

	return net.JoinHostPort(host, strconv.FormatUint(portNum, 10)), nil
}

// parseTargets validates and deduplicates the targets of a scrape request.
// It returns a map of original target (used as label) to dial address, and
// the targets in their original order.
func parseTargets(targets []string) ([]string, map[string]string, error) {
	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("missing 'target' query parameter")
	}
	if len(targets) > maxTargets {
		return nil, nil, fmt.Errorf("too many targets (maximum %d)", maxTargets)
	}

	ordered := make([]string, 0, len(targets))
	addrs := make(map[string]string, len(targets))
	for _, target := range targets {
		if _, seen := addrs[target]; seen {
			continue
		}
		addr, err := parseTarget(target)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid target %q: %w", target, err)
		}
		addrs[target] = addr
		ordered = append(ordered, target)
	}
	return ordered, addrs, nil
}
