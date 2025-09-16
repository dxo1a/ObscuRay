package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/getlantern/systray"
)

var singBoxPath = filepath.Join("assets", "sing-box.exe")
var currentProcess *os.Process

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
			"servers": []interface{}{
				map[string]interface{}{
					"type":            "https",
					"server":          "dns.google",
					"domain_resolver": "dns-local",
					"detour":          "proxy",
					"tag":             "dns-remote",
				},
				map[string]interface{}{
					"type":            "https",
					"server":          "doh.pub",
					"domain_resolver": "dns-local",
					"detour":          "direct",
					"tag":             "dns-direct",
				},
				map[string]interface{}{
					"type":   "local",
					"detour": "direct",
					"tag":    "dns-local",
				},
			},
			"rules": []interface{}{
				map[string]interface{}{
					"domain":         []string{server},
					"domain_keyword": []interface{}{},
					"domain_regex":   []interface{}{},
					"domain_suffix":  []interface{}{},
					"rule_set":       []interface{}{},
					"server":         "dns-direct",
				},
				map[string]interface{}{
					"query_type": []int{32, 33}, // A и AAAA
					"action":     "predefined",
					"rcode":      "NOERROR",
				},
				map[string]interface{}{
					"domain_suffix": ".lan",
					"action":        "predefined",
					"rcode":         "NOERROR",
				},
			},
		},
		"inbounds": []interface{}{
			map[string]interface{}{
				"listen":      "127.0.0.1",
				"listen_port": proxyPort,
				"tag":         "mixed-in",
				"type":        "mixed",
			},
		},
		"log": map[string]interface{}{"level": "info"},
		"outbounds": []interface{}{
			map[string]interface{}{
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
				"multiplex": map[string]interface{}{
					"enabled":     false,
					"protocol":    "h2mux",
					"max_streams": 32,
				},
				"domain_resolver": map[string]interface{}{
					"server": "dns-direct",
				},
			},
			map[string]interface{}{
				"tag":  "direct",
				"type": "direct",
				"domain_resolver": map[string]interface{}{
					"server": "dns-direct",
				},
			},
			map[string]interface{}{
				"tag":  "bypass",
				"type": "direct",
				"domain_resolver": map[string]interface{}{
					"server": "dns-direct",
				},
			},
		},
		"route": map[string]interface{}{
			"auto_detect_interface": false,
			"final":                 "proxy",
			"rule_set": []interface{}{
				map[string]interface{}{
					"tag":             "geosite-cn",
					"type":            "remote",
					"format":          "binary",
					"url":             "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-cn.srs",
					"download_detour": "proxy",
				},
				map[string]interface{}{
					"tag":             "geoip-cn",
					"type":            "remote",
					"format":          "binary",
					"url":             "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-cn.srs",
					"download_detour": "proxy",
				},
			},
			"rules": []interface{}{
				map[string]interface{}{
					"inbound": "mixed-in",
					"action":  "sniff",
				},
				map[string]interface{}{
					"protocol": "dns",
					"action":   "hijack-dns",
				},
				map[string]interface{}{
					"network": "udp",
					"port":    []int{135, 137, 138, 139, 5353},
					"action":  "reject",
				},
				map[string]interface{}{
					"ip_cidr": []string{"224.0.0.0/3", "ff00::/8"},
					"action":  "reject",
				},
				map[string]interface{}{
					"source_ip_cidr": []string{"224.0.0.0/3", "ff00::/8"},
					"action":         "reject",
				},
				map[string]interface{}{
					"ip_is_private": true,
					"outbound":      "direct",
				},
				map[string]interface{}{
					"rule_set": "geoip-cn",
					"outbound": "direct",
				},
			},
		},
		"experimental": map[string]interface{}{
			"cache_file": map[string]interface{}{
				"enabled": true,
			},
			"clash_api": map[string]interface{}{
				"external_controller": "127.0.0.1:9090",
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

	// If type=ws or grpc then add transport
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

func StartProfile(id string) error {
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

	StopProfile()

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
	for i := range Profiles {
		Profiles[i].IsActive = Profiles[i].ID == id
	}
	if err := saveProfiles(); err != nil {
		log.Printf("Failed to save profiles: %v", err)
		return err
	}

	if RunningIconData != nil {
		systray.SetIcon(RunningIconData)
	} else {
		log.Println("Running icon data not set")
	}

	log.Println("Profile started successfully")
	return nil
}

func StopProfile() error {
	log.Println("Stopping profile...")
	if currentProcess != nil {
		log.Println("Killing sing-box process...")
		if err := currentProcess.Kill(); err != nil {
			log.Printf("Failed to kill sing-box: %v", err)
		}
		currentProcess = nil
	}
	for i := range Profiles {
		Profiles[i].IsActive = false
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

	if StoppedIconData != nil {
		systray.SetIcon(StoppedIconData)
	} else {
		log.Println("Stopped icon data not set")
	}

	log.Println("Profile stopped successfully")
	return nil
}

func GetTrafficStats() (map[string]int64, error) {
	if currentProcess == nil {
		return map[string]int64{"download": 0, "upload": 0}, nil
	}

	resp, err := http.Get("http://127.0.0.1:9090/traffic")
	if err != nil {
		log.Printf("Failed to query Clash API: %v", err)
		return nil, fmt.Errorf("failed to query Clash API: %v", err)
	}
	defer resp.Body.Close()

	var stats struct {
		Up   int64 `json:"up"`
		Down int64 `json:"down"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("Failed to decode Clash API response: %v", err)
		return nil, fmt.Errorf("failed to decode Clash API response: %v", err)
	}

	result := map[string]int64{
		"download": stats.Down,
		"upload":   stats.Up,
	}

	return result, nil
}
