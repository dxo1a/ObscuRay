package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

func isValidVLESS(vless string) bool {
	return strings.HasPrefix(vless, "vless://")
}

func parseVLESSName(vless string) string {
	u, err := url.Parse(vless)
	if err != nil || u.Fragment == "" {
		return "Unnamed"
	}
	return u.Fragment
}

func parseVLESS(vless string) (server, port, uuid, flow, network, packetEncoding, security, fingerprint, publicKey, serverName, shortID, path string, err error) {
	u, err := url.Parse(vless)
	if err != nil {
		return "", "", "", "", "", "", "", "", "", "", "", "", fmt.Errorf("invalid VLESS URL: %v", err)
	}
	if u.Scheme != "vless" {
		return "", "", "", "", "", "", "", "", "", "", "", "", fmt.Errorf("not a VLESS URL")
	}

	server = u.Hostname()
	port = u.Port()
	if port == "" {
		port = "443" // По умолчанию для VLESS
	}
	uuid = u.User.Username()
	if uuid == "" {
		return "", "", "", "", "", "", "", "", "", "", "", "", fmt.Errorf("missing UUID in VLESS URL")
	}

	params := u.Query()
	flow = params.Get("flow")
	network = params.Get("type") // tcp/udp/grpc/ws/etc.
	if network == "" {
		network = "tcp"
	}
	packetEncoding = params.Get("packetaddr") // или packetEncoding
	security = params.Get("security")         // none/tls/reality
	if security == "" {
		security = "none"
	}
	fingerprint = params.Get("fp") // utls fingerprint: chrome/firefox/etc.
	publicKey = params.Get("pbk")  // reality public key
	serverName = params.Get("sni") // server name for tls
	shortID = params.Get("sid")    // reality short id
	path = params.Get("spx")       // или path для ws
	if path == "" {
		path = u.Path // Если path указан в URL
	}

	log.Println("Parsed VLESS - server:", server, "port:", port, "uuid:", uuid, "flow:", flow, "network:", network, "security:", security)
	return
}
