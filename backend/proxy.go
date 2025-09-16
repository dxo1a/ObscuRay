package backend

import (
	"fmt"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	proxyPort = 2080
)

func setProxy() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer k.Close()

	if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
		return fmt.Errorf("failed to set ProxyEnable: %v", err)
	}

	proxyServer := fmt.Sprintf("http=127.0.0.1:%s;socks=127.0.0.1:%s", strconv.Itoa(proxyPort), strconv.Itoa(proxyPort))
	if err := k.SetStringValue("ProxyServer", proxyServer); err != nil {
		return fmt.Errorf("failed to set ProxyServer: %v", err)
	}

	if err := k.SetStringValue("ProxyOverride", "<local>"); err != nil {
		return fmt.Errorf("failed to ProxyOverride: %v", err)
	}

	if err := notifyInternetSettingsChanged(); err != nil {
		return fmt.Errorf("failed to notify system: %v", err)
	}

	return nil
}

func unsetProxy() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer k.Close()

	if err := k.SetDWordValue("ProxyEnable", 0); err != nil {
		return fmt.Errorf("failed to set ProxyEnable: %v", err)
	}

	k.DeleteValue("ProxyServer")
	k.DeleteValue("ProxyOverride")

	if err := notifyInternetSettingsChanged(); err != nil {
		return fmt.Errorf("failed to notify system: %v", err)
	}

	return nil
}

func notifyInternetSettingsChanged() error {
	wininet, err := syscall.LoadDLL("wininet.dll")
	if err != nil {
		return fmt.Errorf("failed to load wininet.dll: %v", err)
	}

	internetSetOption, err := wininet.FindProc("InternetSetOptionA")
	if err != nil {
		return fmt.Errorf("failed to find InternetSetOptionA: %v", err)
	}

	// INTERNET_OPTION_SETTINGS_CHANGED (39)
	r1, _, err := internetSetOption.Call(0, 39, 0, 0)
	if r1 == 0 {
		return fmt.Errorf("InternetSetOptionA SETTINGS_CHANGED failed: %v", err)
	}

	// INTERNET_OPTION_REFRESH (37)
	r1, _, err = internetSetOption.Call(0, 37, 0, 0)
	if r1 == 0 {
		return fmt.Errorf("InternetSetOptionA REFRESH failed: %v", err)
	}

	return nil
}
