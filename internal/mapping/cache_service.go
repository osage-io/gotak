package mapping

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// MapCacheService provides offline map tile caching functionality
type MapCacheService struct {
	db         database.DB
	logger     *logger.Logger
	config     *CacheConfig
	httpClient *http.Client
	
	// Download management
	downloadQueue chan *DownloadTask
	workers       int
	stopChan      chan struct{}
	wg            sync.WaitGroup
	
	// Progress tracking
	progressLock sync.RWMutex
	progressMap  map[uuid.UUID]*DownloadProgress
}

// CacheConfig contains configuration for map caching
type CacheConfig struct {
	MaxSizeGB      float64 `yaml:"max_size_gb"`
	ExpirationDays int     `yaml:"expiration_days"`
	CachePath      string  `yaml:"cache_path"`
	Workers        int     `yaml:"workers"`
	TileSources    []TileSource `yaml:"tile_sources"`
}

// TileSource represents a map tile source
type TileSource struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Attribution string `yaml:"attribution"`
	MaxZoom     int    `yaml:"max_zoom"`
	MinZoom     int    `yaml:"min_zoom"`
	Format      string `yaml:"format"` // png, jpg, webp
}

// DownloadTask represents a single tile download task
type DownloadTask struct {
	AreaID uuid.UUID
	Layer  string
	X      int
	Y      int
	Z      int
	URL    string
}

// DownloadProgress tracks download progress for an offline area
type DownloadProgress struct {
	AreaID         uuid.UUID
	TotalTiles     int
	CompletedTiles int
	FailedTiles    int
	StartTime      time.Time
	EstimatedTime  *time.Duration
}

// NewMapCacheService creates a new map cache service
func NewMapCacheService(db database.DB, logger *logger.Logger, config *CacheConfig) *MapCacheService {
	if config.Workers == 0 {
		config.Workers = 4 // Default workers
	}
	
	// Ensure cache directory exists
	os.MkdirAll(config.CachePath, 0755)
	
	mcs := &MapCacheService{
		db:            db,
		logger:        logger,
		config:        config,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		downloadQueue: make(chan *DownloadTask, 1000),
		workers:       config.Workers,
		stopChan:      make(chan struct{}),
		progressMap:   make(map[uuid.UUID]*DownloadProgress),
	}
	
	// Start download workers
	mcs.startWorkers()
	
	return mcs
}

// Start begins the cache service
func (mcs *MapCacheService) Start() error {
	mcs.logger.Info().
		Str("cache_path", mcs.config.CachePath).
		Int("workers", mcs.workers).
		Float64("max_size_gb", mcs.config.MaxSizeGB).
		Msg("Starting map cache service")
	
	// Clean up expired tiles
	go mcs.cleanupExpiredTiles()
	
	return nil
}

// Stop stops the cache service
func (mcs *MapCacheService) Stop() error {
	mcs.logger.Info().Msg("Stopping map cache service")
	
	close(mcs.stopChan)
	mcs.wg.Wait()
	
	return nil
}

// CreateOfflineArea creates a new offline area and starts downloading tiles
func (mcs *MapCacheService) CreateOfflineArea(ctx context.Context, req *CreateOfflineAreaRequest, userID uuid.UUID) (*OfflineArea, error) {
	// Validate zoom levels
	if req.MinZoom > req.MaxZoom || req.MinZoom < 1 || req.MaxZoom > 20 {
		return nil, fmt.Errorf("invalid zoom levels: min=%d, max=%d", req.MinZoom, req.MaxZoom)
	}
	
	// Calculate total tiles and estimated size
	totalTiles := mcs.calculateTileCount(req.Bounds, req.MinZoom, req.MaxZoom, len(req.Layers))
	estimatedSizeMB := float64(totalTiles) * 15.0 / 1024.0 // ~15KB per tile average
	
	// Check storage limits
	if estimatedSizeMB > mcs.config.MaxSizeGB*1024 {
		return nil, fmt.Errorf("offline area too large: %.2f MB (limit: %.2f GB)", 
			estimatedSizeMB, mcs.config.MaxSizeGB)
	}
	
	area := &OfflineArea{
		ID:        uuid.New(),
		Name:      req.Name,
		Bounds:    req.Bounds,
		MinZoom:   req.MinZoom,
		MaxZoom:   req.MaxZoom,
		Layers:    req.Layers,
		Status:    CacheStatusPending,
		Progress:  0,
		SizeMB:    estimatedSizeMB,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save to database
	if err := mcs.saveOfflineAreaToDatabase(ctx, area); err != nil {
		return nil, fmt.Errorf("failed to save offline area: %w", err)
	}
	
	// Start download process
	go mcs.downloadOfflineArea(area)
	
	mcs.logger.Info().
		Str("area_id", area.ID.String()).
		Str("name", area.Name).
		Int("total_tiles", totalTiles).
		Float64("estimated_mb", estimatedSizeMB).
		Msg("Created offline area")
	
	return area, nil
}

// GetOfflineArea retrieves an offline area by ID
func (mcs *MapCacheService) GetOfflineArea(ctx context.Context, areaID uuid.UUID) (*OfflineArea, error) {
	return mcs.getOfflineAreaFromDatabase(ctx, areaID)
}

// ListOfflineAreas returns all offline areas
func (mcs *MapCacheService) ListOfflineAreas(ctx context.Context, limit, offset int) ([]*OfflineArea, error) {
	return mcs.listOfflineAreasFromDatabase(ctx, limit, offset)
}

// DeleteOfflineArea removes an offline area and its cached tiles
func (mcs *MapCacheService) DeleteOfflineArea(ctx context.Context, areaID uuid.UUID) error {
	area, err := mcs.GetOfflineArea(ctx, areaID)
	if err != nil {
		return err
	}
	
	// Remove cached tiles from filesystem
	if err := mcs.removeCachedTiles(area); err != nil {
		mcs.logger.Warn().Err(err).Msg("Failed to remove cached tiles")
	}
	
	// Remove from database
	if err := mcs.deleteOfflineAreaFromDatabase(ctx, areaID); err != nil {
		return err
	}
	
	// Remove from progress tracking
	mcs.progressLock.Lock()
	delete(mcs.progressMap, areaID)
	mcs.progressLock.Unlock()
	
	mcs.logger.Info().
		Str("area_id", areaID.String()).
		Msg("Deleted offline area")
	
	return nil
}

// GetDownloadProgress returns download progress for an offline area
func (mcs *MapCacheService) GetDownloadProgress(areaID uuid.UUID) *DownloadProgress {
	mcs.progressLock.RLock()
	defer mcs.progressLock.RUnlock()
	
	if progress, exists := mcs.progressMap[areaID]; exists {
		// Create a copy to avoid race conditions
		progressCopy := *progress
		return &progressCopy
	}
	
	return nil
}

// GetTile retrieves a cached tile or downloads it if not available
func (mcs *MapCacheService) GetTile(layer string, z, x, y int) ([]byte, error) {
	// Check if tile exists in cache
	tilePath := mcs.getTilePath(layer, z, x, y)
	if data, err := os.ReadFile(tilePath); err == nil {
		// Check if tile is not expired
		if stat, err := os.Stat(tilePath); err == nil {
			expirationTime := time.Duration(mcs.config.ExpirationDays) * 24 * time.Hour
			if time.Since(stat.ModTime()) < expirationTime {
				return data, nil
			}
		}
	}
	
	// Download tile if not cached or expired
	return mcs.downloadSingleTile(layer, z, x, y)
}

// downloadOfflineArea downloads all tiles for an offline area
func (mcs *MapCacheService) downloadOfflineArea(area *OfflineArea) {
	ctx := context.Background()
	
	// Update status to downloading
	mcs.updateOfflineAreaStatus(ctx, area.ID, CacheStatusDownloading)
	
	// Calculate total tiles
	totalTiles := mcs.calculateTileCount(area.Bounds, area.MinZoom, area.MaxZoom, len(area.Layers))
	
	// Initialize progress tracking
	progress := &DownloadProgress{
		AreaID:         area.ID,
		TotalTiles:     totalTiles,
		CompletedTiles: 0,
		FailedTiles:    0,
		StartTime:      time.Now(),
	}
	
	mcs.progressLock.Lock()
	mcs.progressMap[area.ID] = progress
	mcs.progressLock.Unlock()
	
	// Generate download tasks
	tasks := mcs.generateDownloadTasks(area)
	
	mcs.logger.Info().
		Str("area_id", area.ID.String()).
		Int("total_tasks", len(tasks)).
		Msg("Starting offline area download")
	
	// Queue download tasks
	for _, task := range tasks {
		select {
		case mcs.downloadQueue <- task:
		case <-mcs.stopChan:
			return
		}
	}
	
	// Wait for completion or failure
	mcs.waitForDownloadCompletion(area.ID, totalTiles)
}

// generateDownloadTasks creates download tasks for an offline area
func (mcs *MapCacheService) generateDownloadTasks(area *OfflineArea) []*DownloadTask {
	var tasks []*DownloadTask
	
	for _, layer := range area.Layers {
		tileSource := mcs.getTileSource(layer)
		if tileSource == nil {
			mcs.logger.Warn().Str("layer", layer).Msg("Unknown tile source")
			continue
		}
		
		for z := area.MinZoom; z <= area.MaxZoom; z++ {
			bounds := mcs.calculateTileBounds(area.Bounds, z)
			
			for x := bounds.MinX; x <= bounds.MaxX; x++ {
				for y := bounds.MinY; y <= bounds.MaxY; y++ {
					url := mcs.buildTileURL(tileSource.URL, z, x, y)
					
					task := &DownloadTask{
						AreaID: area.ID,
						Layer:  layer,
						X:      x,
						Y:      y,
						Z:      z,
						URL:    url,
					}
					
					tasks = append(tasks, task)
				}
			}
		}
	}
	
	return tasks
}

// TileBounds represents the tile coordinate bounds for a zoom level
type TileBounds struct {
	MinX, MaxX int
	MinY, MaxY int
}

// calculateTileBounds calculates tile bounds for a geographic bounding box at a zoom level
func (mcs *MapCacheService) calculateTileBounds(bounds BoundingBox, zoom int) TileBounds {
	// Convert lat/lng to tile coordinates
	minX := mcs.lngToTileX(bounds.West, zoom)
	maxX := mcs.lngToTileX(bounds.East, zoom)
	minY := mcs.latToTileY(bounds.North, zoom) // Note: Y is flipped
	maxY := mcs.latToTileY(bounds.South, zoom)
	
	return TileBounds{
		MinX: minX,
		MaxX: maxX,
		MinY: minY,
		MaxY: maxY,
	}
}

// lngToTileX converts longitude to tile X coordinate
func (mcs *MapCacheService) lngToTileX(lng float64, zoom int) int {
	return int((lng + 180.0) / 360.0 * float64(int(1)<<uint(zoom)))
}

// latToTileY converts latitude to tile Y coordinate
func (mcs *MapCacheService) latToTileY(lat float64, zoom int) int {
	latRad := lat * 3.14159265359 / 180.0
	return int((1.0 - (math.Log(math.Tan(latRad)+(1.0/math.Cos(latRad)))/3.14159265359)) / 2.0 * float64(int(1)<<uint(zoom)))
}

// calculateTileCount calculates the total number of tiles needed
func (mcs *MapCacheService) calculateTileCount(bounds BoundingBox, minZoom, maxZoom, layerCount int) int {
	totalTiles := 0
	
	for z := minZoom; z <= maxZoom; z++ {
		tileBounds := mcs.calculateTileBounds(bounds, z)
		tilesAtZoom := (tileBounds.MaxX - tileBounds.MinX + 1) * (tileBounds.MaxY - tileBounds.MinY + 1)
		totalTiles += tilesAtZoom * layerCount
	}
	
	return totalTiles
}

// startWorkers starts the download worker goroutines
func (mcs *MapCacheService) startWorkers() {
	for i := 0; i < mcs.workers; i++ {
		mcs.wg.Add(1)
		go mcs.downloadWorker(i)
	}
}

// downloadWorker processes download tasks from the queue
func (mcs *MapCacheService) downloadWorker(workerID int) {
	defer mcs.wg.Done()
	
	mcs.logger.Debug().Int("worker_id", workerID).Msg("Starting download worker")
	
	for {
		select {
		case task := <-mcs.downloadQueue:
			mcs.processDownloadTask(task)
		case <-mcs.stopChan:
			mcs.logger.Debug().Int("worker_id", workerID).Msg("Stopping download worker")
			return
		}
	}
}

// processDownloadTask downloads and caches a single tile
func (mcs *MapCacheService) processDownloadTask(task *DownloadTask) {
	// Check if tile already exists
	tilePath := mcs.getTilePath(task.Layer, task.Z, task.X, task.Y)
	if _, err := os.Stat(tilePath); err == nil {
		mcs.updateProgress(task.AreaID, true)
		return
	}
	
	// Download tile
	resp, err := mcs.httpClient.Get(task.URL)
	if err != nil {
		mcs.logger.Error().Err(err).
			Str("url", task.URL).
			Msg("Failed to download tile")
		mcs.updateProgress(task.AreaID, false)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		mcs.logger.Warn().Int("status", resp.StatusCode).
			Str("url", task.URL).
			Msg("Non-200 response for tile download")
		mcs.updateProgress(task.AreaID, false)
		return
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(tilePath), 0755); err != nil {
		mcs.logger.Error().Err(err).Msg("Failed to create tile directory")
		mcs.updateProgress(task.AreaID, false)
		return
	}
	
	// Save tile to filesystem
	file, err := os.Create(tilePath)
	if err != nil {
		mcs.logger.Error().Err(err).Msg("Failed to create tile file")
		mcs.updateProgress(task.AreaID, false)
		return
	}
	defer file.Close()
	
	if _, err := io.Copy(file, resp.Body); err != nil {
		mcs.logger.Error().Err(err).Msg("Failed to write tile data")
		mcs.updateProgress(task.AreaID, false)
		return
	}
	
	mcs.updateProgress(task.AreaID, true)
}

// updateProgress updates download progress for an offline area
func (mcs *MapCacheService) updateProgress(areaID uuid.UUID, success bool) {
	mcs.progressLock.Lock()
	defer mcs.progressLock.Unlock()
	
	progress, exists := mcs.progressMap[areaID]
	if !exists {
		return
	}
	
	if success {
		progress.CompletedTiles++
	} else {
		progress.FailedTiles++
	}
	
	// Calculate progress percentage
	totalProcessed := progress.CompletedTiles + progress.FailedTiles
	progressPercent := float64(totalProcessed) / float64(progress.TotalTiles) * 100
	
	// Update database with progress
	ctx := context.Background()
	mcs.updateOfflineAreaProgress(ctx, areaID, progressPercent)
	
	// Estimate completion time
	if totalProcessed > 0 {
		elapsed := time.Since(progress.StartTime)
		remaining := float64(progress.TotalTiles-totalProcessed) * elapsed.Seconds() / float64(totalProcessed)
		estimatedTime := time.Duration(remaining) * time.Second
		progress.EstimatedTime = &estimatedTime
	}
}

// waitForDownloadCompletion waits for download to complete
func (mcs *MapCacheService) waitForDownloadCompletion(areaID uuid.UUID, totalTiles int) {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			progress := mcs.GetDownloadProgress(areaID)
			if progress == nil {
				return
			}
			
			totalProcessed := progress.CompletedTiles + progress.FailedTiles
			if totalProcessed >= totalTiles {
				// Download complete
				if progress.FailedTiles == 0 {
					mcs.updateOfflineAreaStatus(ctx, areaID, CacheStatusComplete)
					mcs.logger.Info().
						Str("area_id", areaID.String()).
						Int("tiles", progress.CompletedTiles).
						Msg("Offline area download completed successfully")
				} else {
					mcs.updateOfflineAreaStatus(ctx, areaID, CacheStatusError)
					mcs.logger.Warn().
						Str("area_id", areaID.String()).
						Int("completed", progress.CompletedTiles).
						Int("failed", progress.FailedTiles).
						Msg("Offline area download completed with errors")
				}
				return
			}
		case <-mcs.stopChan:
			return
		}
	}
}

// Helper methods for tile management

func (mcs *MapCacheService) getTilePath(layer string, z, x, y int) string {
	return filepath.Join(mcs.config.CachePath, layer, fmt.Sprintf("%d", z), fmt.Sprintf("%d", x), fmt.Sprintf("%d.png", y))
}

func (mcs *MapCacheService) getTileSource(layerID string) *TileSource {
	for _, source := range mcs.config.TileSources {
		if source.ID == layerID {
			return &source
		}
	}
	return nil
}

func (mcs *MapCacheService) buildTileURL(template string, z, x, y int) string {
	// Replace placeholders in URL template
	// Example: "https://tile.server.com/{z}/{x}/{y}.png"
	url := strings.ReplaceAll(template, "{z}", fmt.Sprintf("%d", z))
	url = strings.ReplaceAll(url, "{x}", fmt.Sprintf("%d", x))
	url = strings.ReplaceAll(url, "{y}", fmt.Sprintf("%d", y))
	return url
}

func (mcs *MapCacheService) downloadSingleTile(layer string, z, x, y int) ([]byte, error) {
	tileSource := mcs.getTileSource(layer)
	if tileSource == nil {
		return nil, fmt.Errorf("unknown tile source: %s", layer)
	}
	
	url := mcs.buildTileURL(tileSource.URL, z, x, y)
	
	resp, err := mcs.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	return io.ReadAll(resp.Body)
}

func (mcs *MapCacheService) removeCachedTiles(area *OfflineArea) error {
	for _, layer := range area.Layers {
		layerPath := filepath.Join(mcs.config.CachePath, layer)
		// Note: This is a simplified removal - in practice, you'd want to
		// only remove tiles that are specifically part of this offline area
		if err := os.RemoveAll(layerPath); err != nil {
			return err
		}
	}
	return nil
}

func (mcs *MapCacheService) cleanupExpiredTiles() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily cleanup
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mcs.performCleanup()
		case <-mcs.stopChan:
			return
		}
	}
}

func (mcs *MapCacheService) performCleanup() {
	// Walk through cache directory and remove expired tiles
	// Implementation depends on your specific cleanup requirements
	mcs.logger.Info().Msg("Starting tile cache cleanup")
	
	expirationTime := time.Duration(mcs.config.ExpirationDays) * 24 * time.Hour
	now := time.Now()
	
	filepath.Walk(mcs.config.CachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		
		if !info.IsDir() && now.Sub(info.ModTime()) > expirationTime {
			if err := os.Remove(path); err != nil {
				mcs.logger.Warn().Err(err).Str("path", path).Msg("Failed to remove expired tile")
			}
		}
		
		return nil
	})
	
	mcs.logger.Info().Msg("Tile cache cleanup completed")
}

// Database operations
func (mcs *MapCacheService) saveOfflineAreaToDatabase(ctx context.Context, area *OfflineArea) error {
	// Implementation depends on your database schema
	return nil
}

func (mcs *MapCacheService) getOfflineAreaFromDatabase(ctx context.Context, areaID uuid.UUID) (*OfflineArea, error) {
	// Implementation depends on your database schema
	return nil, nil
}

func (mcs *MapCacheService) listOfflineAreasFromDatabase(ctx context.Context, limit, offset int) ([]*OfflineArea, error) {
	// Implementation depends on your database schema
	return nil, nil
}

func (mcs *MapCacheService) deleteOfflineAreaFromDatabase(ctx context.Context, areaID uuid.UUID) error {
	// Implementation depends on your database schema
	return nil
}

func (mcs *MapCacheService) updateOfflineAreaStatus(ctx context.Context, areaID uuid.UUID, status CacheStatus) error {
	// Implementation depends on your database schema
	return nil
}

func (mcs *MapCacheService) updateOfflineAreaProgress(ctx context.Context, areaID uuid.UUID, progress float64) error {
	// Implementation depends on your database schema
	return nil
}
