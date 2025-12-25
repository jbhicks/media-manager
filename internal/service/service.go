package service

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

type MediaService struct {
	config          *config.Config
	db              *db.Database
	downloadManager *DownloadManager
	httpServer      *HTTPServer
	stopChan        chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.Mutex
}

func NewMediaService(cfg *config.Config, database *db.Database) (*MediaService, error) {
	return &MediaService{
		config:   cfg,
		db:       database,
		stopChan: make(chan struct{}),
		running:  false,
	}, nil
}

func (s *MediaService) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("service is already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Println("[SERVICE] Media Manager service starting...")

	serviceConfig, err := s.getOrCreateServiceConfig()
	if err != nil {
		return fmt.Errorf("failed to load service config: %w", err)
	}

	if !serviceConfig.DownloadEnabled {
		log.Println("[SERVICE] Download functionality is disabled in configuration")
	}

	downloadManager, err := NewDownloadManager(s.db, serviceConfig)
	if err != nil {
		return fmt.Errorf("failed to create download manager: %w", err)
	}
	s.downloadManager = downloadManager

	startPort := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			startPort = p
		}
	}

	port := s.findAvailablePort(startPort)
	httpServer := NewHTTPServer(port, s.db, downloadManager)
	if err := httpServer.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	s.httpServer = httpServer

	s.wg.Add(1)
	go s.runScheduler(serviceConfig)

	s.wg.Add(1)
	go s.monitorDownloads()

	log.Println("[SERVICE] Media Manager service started successfully")
	return nil
}

func (s *MediaService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	log.Println("[SERVICE] Stopping Media Manager service...")

	if s.httpServer != nil {
		s.httpServer.Stop()
	}

	close(s.stopChan)
	s.wg.Wait()
	log.Println("[SERVICE] Media Manager service stopped")
}

func (s *MediaService) Wait() {
	s.wg.Wait()
}

func (s *MediaService) runScheduler(cfg *models.ServiceConfig) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(cfg.ScheduleInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("[SCHEDULER] Starting scheduler with interval: %d seconds", cfg.ScheduleInterval)

	for {
		select {
		case <-s.stopChan:
			log.Println("[SCHEDULER] Scheduler stopped")
			return
		case <-ticker.C:
			if cfg.DownloadEnabled {
				s.checkDownloadRules()
			}
		}
	}
}

func (s *MediaService) checkDownloadRules() {
	log.Println("[SCHEDULER] Checking download rules...")

	var rules []models.DownloadRule
	if err := s.db.GetDB().Where("enabled = ?", true).Find(&rules).Error; err != nil {
		log.Printf("[ERROR] Failed to fetch download rules: %v", err)
		return
	}

	log.Printf("[SCHEDULER] Found %d enabled download rules", len(rules))

	for _, rule := range rules {
		if rule.AutoDownload {
			log.Printf("[SCHEDULER] Processing auto-download rule: %s", rule.Name)
			s.processDownloadRule(&rule)
		}
	}
}

func (s *MediaService) processDownloadRule(rule *models.DownloadRule) {
	log.Printf("[RULE] Processing rule: %s (Query: %s)", rule.Name, rule.SearchQuery)

	if err := s.downloadManager.SearchAndDownload(rule); err != nil {
		log.Printf("[ERROR] Failed to process download rule: %v", err)
	}

	s.db.GetDB().Model(rule).Update("last_run", time.Now())
}

func (s *MediaService) getOrCreateServiceConfig() (*models.ServiceConfig, error) {
	var cfg models.ServiceConfig
	result := s.db.GetDB().First(&cfg)

	if result.Error != nil {
		cfg = models.ServiceConfig{
			DownloadEnabled:        false,
			ScheduleInterval:       3600,
			MaxConcurrentDownloads: 5,
			TorrentClientType:      "transmission",
			TorrentClientHost:      "localhost:9091",
			JellyfinURL:            "http://localhost:8096",
			JellyfinAPIKey:         "",
		}
		if err := s.db.GetDB().Create(&cfg).Error; err != nil {
			return nil, err
		}
		log.Println("[SERVICE] Created default service configuration")
	}

	return &cfg, nil
}

func (s *MediaService) findAvailablePort(startPort int) string {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			log.Printf("[SERVICE] Found available port: %d", port)
			return addr
		}
	}
	log.Printf("[SERVICE] No available port found, using default :%d", startPort)
	return fmt.Sprintf(":%d", startPort)
}

func (s *MediaService) monitorDownloads() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("[MONITOR] Starting download progress monitor")

	for {
		select {
		case <-s.stopChan:
			log.Println("[MONITOR] Download monitor stopped")
			return
		case <-ticker.C:
			if s.downloadManager != nil {
				s.downloadManager.ProcessPendingDownloads()
				s.downloadManager.UpdateAllProgress()
			}
		}
	}
}
