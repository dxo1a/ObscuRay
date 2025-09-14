package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

var singBoxPath = filepath.Join("assets", "sing-box.exe")
var currentProcess *os.Process

// var geoDir = filepath.Join(os.TempDir(), "obscuray-geo")
var geoIPPath = filepath.Join("assets", "geoip.db")
var geoSitePath = filepath.Join("assets", "geosite.db")

// func downloadGeoFiles() error {
// 	log.Println("Checking geo files...")
// 	os.MkdirAll(geoDir, 0755)
// 	if _, err := os.Stat(geoIPPath); os.IsNotExist(err) {
// 		log.Println("Downloading geoip.dat...")
// 		tempPath := filepath.Join(geoDir, "geoip.temp")
// 		if err := downloadFile("https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat", tempPath); err != nil {
// 			log.Printf("Failed to download geoip.dat: %v", err)
// 			return err
// 		}
// 		if err := os.Rename(tempPath, geoIPPath); err != nil {
// 			log.Printf("Failed to rename geoip.temp: %v", err)
// 			return err
// 		}
// 	}
// 	if _, err := os.Stat(geoSitePath); os.IsNotExist(err) {
// 		log.Println("Downloading geosite.dat...")
// 		tempPath := filepath.Join(geoDir, "geosite.temp")
// 		if err := downloadFile("https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat", tempPath); err != nil {
// 			log.Printf("Failed to download geosite.dat: %v", err)
// 			return err
// 		}
// 		if err := os.Rename(tempPath, geoSitePath); err != nil {
// 			log.Printf("Failed to rename geosite.temp: %v", err)
// 			return err
// 		}
// 	}
// 	log.Println("Geo files ready")
// 	return nil
// }

// func downloadFile(url string, path string) error {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	f, err := os.Create(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	_, err = io.Copy(f, resp.Body)
// 	return err
// }

func generateConfig(vless string) (string, error) {
	log.Println("Generating config for VLESS: ", vless)
	server, portStr, uuid, flow, network, packetEncoding, security, fingerprint, publicKey, serverName, shortID, path, err := parseVLESS(vless)
	if err != nil {
		log.Printf("Failed to parse VLESS: %v", err)
		return "", err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid port: %v", err)
		return "", fmt.Errorf("invalid port: %v", err)
	}

	if packetEncoding == "" {
		packetEncoding = ""
	}

	config := map[string]interface{}{
		"dns": map[string]interface{}{
			"independent_cache": true,
			"rules": []interface{}{
				map[string]interface{}{"outbound": "any", "server": "dns-direct"},
				map[string]interface{}{
					"domain":         []string{server},
					"domain_keyword": []interface{}{},
					"domain_regex":   []interface{}{},
					"domain_suffix":  []interface{}{},
					"geosite":        []interface{}{},
					"server":         "dns-direct",
				},
				map[string]interface{}{"query_type": []int{32, 33}, "server": "dns-block"},
				map[string]interface{}{"domain_suffix": ".lan", "server": "dns-block"},
			},
			"servers": []interface{}{
				map[string]interface{}{"address": "https://dns.google/dns-query", "address_resolver": "dns-local", "detour": "proxy", "strategy": "", "tag": "dns-remote"},
				map[string]interface{}{"address": "https://doh.pub/dns-query", "address_resolver": "dns-local", "detour": "direct", "strategy": "", "tag": "dns-direct"},
				map[string]interface{}{"address": "rcode://success", "tag": "dns-block"},
				map[string]interface{}{"address": "local", "detour": "direct", "tag": "dns-local"},
			},
		},
		"inbounds": []interface{}{
			map[string]interface{}{
				"domain_strategy":            "",
				"listen":                     "127.0.0.1",
				"listen_port":                proxyPort,
				"sniff":                      true,
				"sniff_override_destination": false,
				"tag":                        "mixed-in",
				"type":                       "mixed",
			},
		},
		"log": map[string]interface{}{"level": "info"},
		"outbounds": []interface{}{
			map[string]interface{}{
				"domain_strategy": "",
				"flow":            flow,
				"packet_encoding": packetEncoding,
				"server":          server,
				"server_port":     port,
				"tag":             "proxy",
				"tls": map[string]interface{}{
					"enabled": true,
					"utls": map[string]interface{}{
						"enabled":     true,
						"fingerprint": fingerprint,
					},
				},
				"type": "vless",
				"uuid": uuid,
			},
			map[string]interface{}{"tag": "direct", "type": "direct"},
			map[string]interface{}{"tag": "bypass", "type": "direct"},
			map[string]interface{}{"tag": "block", "type": "block"},
			map[string]interface{}{"tag": "dns-out", "type": "dns"},
		},
		"route": map[string]interface{}{
			"auto_detect_interface": false,
			"final":                 "proxy",
			"geoip":                 map[string]interface{}{"path": "assets/geoip.db"},
			"geosite":               map[string]interface{}{"path": "assets/geosite.db"},
			"rules": []interface{}{
				map[string]interface{}{"outbound": "dns-out", "protocol": "dns"},
				map[string]interface{}{"network": "udp", "outbound": "block", "port": []int{135, 137, 138, 139, 5353}},
				map[string]interface{}{"ip_cidr": []string{"224.0.0.0/3", "ff00::/8"}, "outbound": "block"},
				map[string]interface{}{"outbound": "block", "source_ip_cidr": []string{"224.0.0.0/3", "ff00::/8"}},
			},
		},
	}

	outbound := config["outbounds"].([]interface{})[0].(map[string]interface{})
	tls := outbound["tls"].(map[string]interface{})
	if security == "reality" && publicKey != "" {
		tls["reality"] = map[string]interface{}{
			"enabled":    true,
			"public_key": publicKey,
			"short_id":   shortID,
		}
	}
	if serverName != "" {
		tls["server_name"] = serverName
	}

	// Network/transport
	outbound["network"] = network

	// If type=ws or grpc then add transport (for example for spx=/ as path)
	if network == "ws" || network == "grpc" {
		transportType := network
		transport := map[string]interface{}{
			"type": transportType,
		}
		if path != "" {
			transport["path"] = path
		}
		outbound["transport"] = transport
	}

	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Printf("Failed to marshal config to JSON: %v", err)
		return "", err
	}
	configFile := filepath.Join(os.TempDir(), "obscuray-config.json")
	err = os.WriteFile(configFile, jsonData, 0644)
	if err != nil {
		log.Printf("Failed to write config file: %v", err)
		return "", err
	}
	log.Println("Config generated at:", configFile)
	return configFile, nil
}

func startProfile(id string) error {
	log.Println("Starting profile with ID:", id)
	profile := findProfile(id)
	if profile == nil {
		log.Println("Profile not found")
		return fmt.Errorf("profile not found")
	}
	if profile.IsActive {
		log.Println("Profile already active")
		return fmt.Errorf("profile already active")
	}

	stopProfile()

	// downloading geo files
	log.Println("Checking geo files...")
	if _, err := os.Stat(geoIPPath); os.IsNotExist(err) {
		log.Println("geoip.db not found in assets")
		return fmt.Errorf("geoip.db not found in assets")
	}
	if _, err := os.Stat(geoSitePath); os.IsNotExist(err) {
		log.Printf("geosite.db not found in assets")
		return fmt.Errorf("geosite.db not found in assets")
	}

	// generate config
	configFile, err := generateConfig(profile.VLESS)
	if err != nil {
		log.Printf("Failed to generate config: %v", err)
		return err
	}

	// start sing-box
	log.Println("Starting sing-box with config:", configFile)
	cmd := exec.Command(singBoxPath, "run", "-c", configFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = os.Stdout // sing-box logs into console
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start sing-box: %v", err)
		return err
	}
	currentProcess = cmd.Process

	// setup proxy
	log.Println("Sing-box started, setting proxy...")
	if err := setProxy(); err != nil {
		log.Printf("Failed to set proxy: %v", err)
		cmd.Process.Kill()
		return err
	}

	// refresh status
	log.Println("Proxy set, updating profiles...")
	for i := range profiles {
		profiles[i].IsActive = profiles[i].ID == id
	}
	if err := saveProfiles(); err != nil {
		log.Printf("Failed to save profiles: %v", err)
		return err
	}
	log.Println("Profile started successfully")
	return nil
}

func stopProfile() error {
	log.Println("Stopping profile...")
	if currentProcess != nil {
		log.Println("Killing sing-box process...")
		if err := currentProcess.Kill(); err != nil {
			log.Printf("Failed to kill sing-box: %v", err)
		}
		currentProcess = nil
	}
	for i := range profiles {
		profiles[i].IsActive = false
	}

	log.Println("Unsetting proxy...")
	if err := unsetProxy(); err != nil {
		log.Printf("Failed to unset proxy: %v", err)
		return err
	}
	log.Println("Saving profiles...")
	if err := saveProfiles(); err != nil {
		log.Printf("Failed to save profiles: %v", err)
		return err
	}
	log.Println("Profile stopped successfully")
	return nil
}
