package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/dfedick/gotak/internal/mapping"
	"github.com/dfedick/gotak/pkg/logger"
)

// MappingHandler handles HTTP requests for mapping functionality
type MappingHandler struct {
	routeService     *mapping.RouteService
	geofenceService  *mapping.GeofenceService
	cacheService     *mapping.MapCacheService
	logger           *logger.Logger
}

// NewMappingHandler creates a new mapping handler
func NewMappingHandler(
	routeService *mapping.RouteService,
	geofenceService *mapping.GeofenceService,
	cacheService *mapping.MapCacheService,
	logger *logger.Logger,
) *MappingHandler {
	return &MappingHandler{
		routeService:    routeService,
		geofenceService: geofenceService,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// RegisterRoutes registers mapping-related routes with the Gin router
func (h *MappingHandler) RegisterRoutes(r *gin.RouterGroup) {
	mapping := r.Group("/mapping")
	{
		// Route management
		routes := mapping.Group("/routes")
		{
			routes.POST("", h.CreateRoute)
			routes.GET("", h.ListRoutes)
			routes.GET("/:id", h.GetRoute)
			routes.PUT("/:id", h.UpdateRoute)
			routes.DELETE("/:id", h.DeleteRoute)
			routes.POST("/:id/recalculate", h.RecalculateRoute)
		}

		// Geofence management
		geofences := mapping.Group("/geofences")
		{
			geofences.POST("", h.CreateGeofence)
			geofences.GET("", h.ListGeofences)
			geofences.GET("/:id", h.GetGeofence)
			geofences.PUT("/:id", h.UpdateGeofence)
			geofences.DELETE("/:id", h.DeleteGeofence)
			geofences.GET("/violations", h.GetViolations)
			geofences.PUT("/violations/:id/acknowledge", h.AcknowledgeViolation)
		}

		// Offline map caching
		offline := mapping.Group("/offline")
		{
			offline.POST("/areas", h.CreateOfflineArea)
			offline.GET("/areas", h.ListOfflineAreas)
			offline.GET("/areas/:id", h.GetOfflineArea)
			offline.DELETE("/areas/:id", h.DeleteOfflineArea)
			offline.GET("/areas/:id/progress", h.GetDownloadProgress)
			offline.GET("/tiles/:layer/:z/:x/:y", h.GetCachedTile)
		}
	}
}

// Route handlers

// CreateRoute creates a new route with waypoints
func (h *MappingHandler) CreateRoute(c *gin.Context) {
	var req mapping.CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid route creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user and group from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	groupID, exists := c.Get("group_id")
	if !exists {
		groupID = "default" // Fallback group
	}

	route, err := h.routeService.CreateRoute(c.Request.Context(), &req, userID.(uuid.UUID), groupID.(string))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create route")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create route"})
		return
	}

	c.JSON(http.StatusCreated, route)
}

// GetRoute retrieves a route by ID
func (h *MappingHandler) GetRoute(c *gin.Context) {
	idStr := c.Param("id")
	routeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	route, err := h.routeService.GetRoute(c.Request.Context(), routeID)
	if err != nil {
		h.logger.Error().Err(err).Str("route_id", idStr).Msg("Failed to get route")
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	c.JSON(http.StatusOK, route)
}

// ListRoutes returns routes for the user's group
func (h *MappingHandler) ListRoutes(c *gin.Context) {
	groupID, exists := c.Get("group_id")
	if !exists {
		groupID = "default"
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	routes, err := h.routeService.ListRoutes(c.Request.Context(), groupID.(string), limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list routes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list routes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"routes": routes})
}

// UpdateRoute updates an existing route
func (h *MappingHandler) UpdateRoute(c *gin.Context) {
	idStr := c.Param("id")
	routeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	route, err := h.routeService.UpdateRoute(c.Request.Context(), routeID, updates)
	if err != nil {
		h.logger.Error().Err(err).Str("route_id", idStr).Msg("Failed to update route")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update route"})
		return
	}

	c.JSON(http.StatusOK, route)
}

// DeleteRoute removes a route
func (h *MappingHandler) DeleteRoute(c *gin.Context) {
	idStr := c.Param("id")
	routeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	if err := h.routeService.DeleteRoute(c.Request.Context(), routeID); err != nil {
		h.logger.Error().Err(err).Str("route_id", idStr).Msg("Failed to delete route")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete route"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted successfully"})
}

// RecalculateRoute recalculates a route with new options
func (h *MappingHandler) RecalculateRoute(c *gin.Context) {
	idStr := c.Param("id")
	routeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	var options mapping.RouteOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route options"})
		return
	}

	route, err := h.routeService.RecalculateRoute(c.Request.Context(), routeID, options)
	if err != nil {
		h.logger.Error().Err(err).Str("route_id", idStr).Msg("Failed to recalculate route")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate route"})
		return
	}

	c.JSON(http.StatusOK, route)
}

// Geofence handlers

// CreateGeofence creates a new geofence
func (h *MappingHandler) CreateGeofence(c *gin.Context) {
	var req mapping.CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid geofence creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	groupID, exists := c.Get("group_id")
	if !exists {
		groupID = "default"
	}

	geofence, err := h.geofenceService.CreateGeofence(c.Request.Context(), &req, userID.(uuid.UUID), groupID.(string))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create geofence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create geofence"})
		return
	}

	c.JSON(http.StatusCreated, geofence)
}

// GetGeofence retrieves a geofence by ID
func (h *MappingHandler) GetGeofence(c *gin.Context) {
	idStr := c.Param("id")
	geofenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	geofence, err := h.geofenceService.GetGeofence(c.Request.Context(), geofenceID)
	if err != nil {
		h.logger.Error().Err(err).Str("geofence_id", idStr).Msg("Failed to get geofence")
		c.JSON(http.StatusNotFound, gin.H{"error": "Geofence not found"})
		return
	}

	c.JSON(http.StatusOK, geofence)
}

// ListGeofences returns geofences for the user's group
func (h *MappingHandler) ListGeofences(c *gin.Context) {
	groupID, exists := c.Get("group_id")
	if !exists {
		groupID = "default"
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	geofences, err := h.geofenceService.ListGeofences(c.Request.Context(), groupID.(string), limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list geofences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list geofences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"geofences": geofences})
}

// UpdateGeofence updates an existing geofence
func (h *MappingHandler) UpdateGeofence(c *gin.Context) {
	idStr := c.Param("id")
	geofenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	updates["updated_at"] = time.Now()

	geofence, err := h.geofenceService.UpdateGeofence(c.Request.Context(), geofenceID, updates)
	if err != nil {
		h.logger.Error().Err(err).Str("geofence_id", idStr).Msg("Failed to update geofence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update geofence"})
		return
	}

	c.JSON(http.StatusOK, geofence)
}

// DeleteGeofence removes a geofence
func (h *MappingHandler) DeleteGeofence(c *gin.Context) {
	idStr := c.Param("id")
	geofenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	if err := h.geofenceService.DeleteGeofence(c.Request.Context(), geofenceID); err != nil {
		h.logger.Error().Err(err).Str("geofence_id", idStr).Msg("Failed to delete geofence")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete geofence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence deleted successfully"})
}

// GetViolations retrieves geofence violations
func (h *MappingHandler) GetViolations(c *gin.Context) {
	var geofenceID *uuid.UUID
	var entityID *string

	if geoIDStr := c.Query("geofence_id"); geoIDStr != "" {
		if id, err := uuid.Parse(geoIDStr); err == nil {
			geofenceID = &id
		}
	}

	if entID := c.Query("entity_id"); entID != "" {
		entityID = &entID
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	violations, err := h.geofenceService.GetViolations(c.Request.Context(), geofenceID, entityID, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get violations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get violations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"violations": violations})
}

// AcknowledgeViolation marks a violation as acknowledged
func (h *MappingHandler) AcknowledgeViolation(c *gin.Context) {
	idStr := c.Param("id")
	violationID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid violation ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if err := h.geofenceService.AcknowledgeViolation(c.Request.Context(), violationID, userID.(uuid.UUID)); err != nil {
		h.logger.Error().Err(err).Str("violation_id", idStr).Msg("Failed to acknowledge violation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge violation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Violation acknowledged successfully"})
}

// Offline map handlers

// CreateOfflineArea creates a new offline area for map caching
func (h *MappingHandler) CreateOfflineArea(c *gin.Context) {
	var req mapping.CreateOfflineAreaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid offline area creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	area, err := h.cacheService.CreateOfflineArea(c.Request.Context(), &req, userID.(uuid.UUID))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create offline area")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offline area"})
		return
	}

	c.JSON(http.StatusCreated, area)
}

// GetOfflineArea retrieves an offline area by ID
func (h *MappingHandler) GetOfflineArea(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	area, err := h.cacheService.GetOfflineArea(c.Request.Context(), areaID)
	if err != nil {
		h.logger.Error().Err(err).Str("area_id", idStr).Msg("Failed to get offline area")
		c.JSON(http.StatusNotFound, gin.H{"error": "Offline area not found"})
		return
	}

	c.JSON(http.StatusOK, area)
}

// ListOfflineAreas returns all offline areas
func (h *MappingHandler) ListOfflineAreas(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	areas, err := h.cacheService.ListOfflineAreas(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list offline areas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list offline areas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"areas": areas})
}

// DeleteOfflineArea removes an offline area and its cached tiles
func (h *MappingHandler) DeleteOfflineArea(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	if err := h.cacheService.DeleteOfflineArea(c.Request.Context(), areaID); err != nil {
		h.logger.Error().Err(err).Str("area_id", idStr).Msg("Failed to delete offline area")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete offline area"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Offline area deleted successfully"})
}

// GetDownloadProgress returns download progress for an offline area
func (h *MappingHandler) GetDownloadProgress(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	progress := h.cacheService.GetDownloadProgress(areaID)
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Download progress not found"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetCachedTile retrieves a cached tile or downloads it if not available
func (h *MappingHandler) GetCachedTile(c *gin.Context) {
	layer := c.Param("layer")
	z, _ := strconv.Atoi(c.Param("z"))
	x, _ := strconv.Atoi(c.Param("x"))
	y, _ := strconv.Atoi(c.Param("y"))

	tileData, err := h.cacheService.GetTile(layer, z, x, y)
	if err != nil {
		h.logger.Error().Err(err).
			Str("layer", layer).
			Int("z", z).Int("x", x).Int("y", y).
			Msg("Failed to get cached tile")
		c.JSON(http.StatusNotFound, gin.H{"error": "Tile not found"})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	c.Data(http.StatusOK, "image/png", tileData)
}

// Helper functions for extracting common parameters
func (h *MappingHandler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

func (h *MappingHandler) getGroupID(c *gin.Context) string {
	groupID, exists := c.Get("group_id")
	if !exists {
		return "default"
	}
	return groupID.(string)
}
