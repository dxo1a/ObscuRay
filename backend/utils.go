package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var Translations = map[string]map[string]string{
	"en": {
		"tooltip":  "ObscuRay VLESS Proxy",
		"show":     "Show",
		"showDesc": "Show application",
		"quit":     "Quit",
		"quitDesc": "Quit application",
	},
	"ru": {
		"tooltip":  "ObscuRay VLESS прокси",
		"show":     "Показать",
		"showDesc": "Показать приложение",
		"quit":     "Выйти",
		"quitDesc": "Выйти из приложения",
	},
}

func GetLang() string {
	lang := os.Getenv("LANG")
	if len(lang) >= 2 {
		if lang[:2] == "ru" {
			return "ru"
		}
	}
	return "en"
}

func SetupLogging() error {
	logDir := filepath.Join(os.Getenv("APPDATA"), "ObscuRay")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	logFilePath := filepath.Join(logDir, "app.log")
	if _, err := os.Stat(logFilePath); err == nil {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		newLogFilePath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", timestamp))
		if err := os.Rename(logFilePath, newLogFilePath); err != nil {
			return fmt.Errorf("failed to rename log file: %v", err)
		}
		log.Printf("Renamed old app.log to %s", newLogFilePath)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	log.SetOutput(logFile)

	CleanOldLogs(logDir)
	return nil
}

func CleanOldLogs(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to read log directory: %v", err)
		return
	}

	var logFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), "app_") && strings.HasSuffix(file.Name(), ".log") {
			logFiles = append(logFiles, file.Name())
		}
	}

	// Сортируем файлы по времени модификации (старые первыми)
	sort.Slice(logFiles, func(i, j int) bool {
		iTime, _ := os.Stat(filepath.Join(dir, logFiles[i]))
		jTime, _ := os.Stat(filepath.Join(dir, logFiles[j]))
		return iTime.ModTime().Before(jTime.ModTime())
	})

	// Удаляем лишние файлы, если их больше 5
	if len(logFiles) > 5 {
		for _, file := range logFiles[:len(logFiles)-5] {
			if err := os.Remove(filepath.Join(dir, file)); err != nil {
				log.Printf("Failed to remove old log file %s: %v", file, err)
			} else {
				log.Printf("Removed old log file: %s", file)
			}
		}
	}
}
