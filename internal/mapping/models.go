package mapping

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Common types for mapping functionality

// Point represents a geographical point
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// BoundingBox represents a geographical bounding box
type BoundingBox struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

// LineString represents a GeoJSON LineString
type LineString struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// Polygon represents a GeoJSON Polygon
type Polygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// Circle represents a geographical circle
type Circle struct {
	Center Point   `json:"center"`
	Radius float64 `json:"radius"` // meters
}

// Route represents a calculated route between waypoints
type Route struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Description string        `json:"description" db:"description"`
	CreatedBy   uuid.UUID     `json:"created_by" db:"created_by"`
	GroupID     string        `json:"group_id" db:"group_id"`
	
	// Route data
	Waypoints   []Waypoint    `json:"waypoints"`
	Geometry    LineString    `json:"geometry"`     // GeoJSON LineString
	Distance    float64       `json:"distance"`     // meters
	Duration    time.Duration `json:"duration"`     // estimated travel time
	
	// Route options
	RouteType   RouteType     `json:"route_type" db:"route_type"`
	Vehicle     VehicleType   `json:"vehicle" db:"vehicle"`
	Optimize    bool          `json:"optimize" db:"optimize"`
	
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// Waypoint represents a point along a route
type Waypoint struct {
	ID          uuid.UUID `json:"id" db:"id"`
	RouteID     uuid.UUID `json:"route_id" db:"route_id"`
	Sequence    int       `json:"sequence" db:"sequence"`
	Lat         float64   `json:"lat" db:"lat"`
	Lng         float64   `json:"lng" db:"lng"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ETA         *time.Time `json:"eta,omitempty" db:"eta"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RouteType represents the type of route calculation
type RouteType string

const (
	RouteTypeFastest  RouteType = "fastest"
	RouteTypeShortest RouteType = "shortest"
	RouteTypeTactical RouteType = "tactical"  // Avoid main roads
	RouteTypeOffRoad  RouteType = "offroad"   // Direct line
)

// Valid returns true if the route type is valid
func (rt RouteType) Valid() bool {
	switch rt {
	case RouteTypeFastest, RouteTypeShortest, RouteTypeTactical, RouteTypeOffRoad:
		return true
	}
	return false
}

// String returns the string representation
func (rt RouteType) String() string {
	return string(rt)
}

// Scan implements the sql.Scanner interface
func (rt *RouteType) Scan(value interface{}) error {
	if value == nil {
		*rt = RouteTypeFastest
		return nil
	}
	if str, ok := value.(string); ok {
		*rt = RouteType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into RouteType", value)
}

// Value implements the driver.Valuer interface
func (rt RouteType) Value() (driver.Value, error) {
	return string(rt), nil
}

// VehicleType represents the vehicle type for route calculation
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeTruck      VehicleType = "truck"
	VehicleTypeBicycle    VehicleType = "bicycle"
	VehicleTypeFoot       VehicleType = "foot"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
)

// Valid returns true if the vehicle type is valid
func (vt VehicleType) Valid() bool {
	switch vt {
	case VehicleTypeCar, VehicleTypeTruck, VehicleTypeBicycle, VehicleTypeFoot, VehicleTypeMotorcycle:
		return true
	}
	return false
}

// String returns the string representation
func (vt VehicleType) String() string {
	return string(vt)
}

// Scan implements the sql.Scanner interface
func (vt *VehicleType) Scan(value interface{}) error {
	if value == nil {
		*vt = VehicleTypeCar
		return nil
	}
	if str, ok := value.(string); ok {
		*vt = VehicleType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into VehicleType", value)
}

// Value implements the driver.Valuer interface
func (vt VehicleType) Value() (driver.Value, error) {
	return string(vt), nil
}

// Geofence represents a geographical boundary for monitoring
type Geofence struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	Type          GeofenceType    `json:"type" db:"type"`
	Geometry      interface{}     `json:"geometry"`    // GeoJSON geometry
	
	// Monitoring settings
	Enabled       bool            `json:"enabled" db:"enabled"`
	AlertOnEnter  bool            `json:"alert_on_enter" db:"alert_on_enter"`
	AlertOnExit   bool            `json:"alert_on_exit" db:"alert_on_exit"`
	
	// Access control
	CreatedBy     uuid.UUID       `json:"created_by" db:"created_by"`
	GroupID       string          `json:"group_id" db:"group_id"`
	
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// GeofenceType represents the type of geofence geometry
type GeofenceType string

const (
	GeofenceTypeCircle    GeofenceType = "circle"
	GeofenceTypePolygon   GeofenceType = "polygon"
	GeofenceTypeRectangle GeofenceType = "rectangle"
)

// Valid returns true if the geofence type is valid
func (gt GeofenceType) Valid() bool {
	switch gt {
	case GeofenceTypeCircle, GeofenceTypePolygon, GeofenceTypeRectangle:
		return true
	}
	return false
}

// String returns the string representation
func (gt GeofenceType) String() string {
	return string(gt)
}

// Scan implements the sql.Scanner interface
func (gt *GeofenceType) Scan(value interface{}) error {
	if value == nil {
		*gt = GeofenceTypeCircle
		return nil
	}
	if str, ok := value.(string); ok {
		*gt = GeofenceType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into GeofenceType", value)
}

// Value implements the driver.Valuer interface
func (gt GeofenceType) Value() (driver.Value, error) {
	return string(gt), nil
}

// GeofenceViolation represents a boundary violation event
type GeofenceViolation struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	GeofenceID     uuid.UUID     `json:"geofence_id" db:"geofence_id"`
	EntityID       string        `json:"entity_id" db:"entity_id"`
	ViolationType  ViolationType `json:"violation_type" db:"violation_type"`
	Position       Point         `json:"position"`
	Timestamp      time.Time     `json:"timestamp" db:"timestamp"`
	Acknowledged   bool          `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy *uuid.UUID    `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
}

// ViolationType represents the type of geofence violation
type ViolationType string

const (
	ViolationEnter ViolationType = "enter"
	ViolationExit  ViolationType = "exit"
)

// Valid returns true if the violation type is valid
func (vt ViolationType) Valid() bool {
	switch vt {
	case ViolationEnter, ViolationExit:
		return true
	}
	return false
}

// String returns the string representation
func (vt ViolationType) String() string {
	return string(vt)
}

// Scan implements the sql.Scanner interface
func (vt *ViolationType) Scan(value interface{}) error {
	if value == nil {
		*vt = ViolationEnter
		return nil
	}
	if str, ok := value.(string); ok {
		*vt = ViolationType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ViolationType", value)
}

// Value implements the driver.Valuer interface
func (vt ViolationType) Value() (driver.Value, error) {
	return string(vt), nil
}

// OfflineArea represents an area for offline map caching
type OfflineArea struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Bounds    BoundingBox `json:"bounds"`
	MinZoom   int         `json:"min_zoom" db:"min_zoom"`
	MaxZoom   int         `json:"max_zoom" db:"max_zoom"`
	Layers    []string    `json:"layers"`
	Status    CacheStatus `json:"status" db:"status"`
	Progress  float64     `json:"progress" db:"progress"`
	SizeMB    float64     `json:"size_mb" db:"size_mb"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// CacheStatus represents the status of offline area caching
type CacheStatus string

const (
	CacheStatusPending     CacheStatus = "pending"
	CacheStatusDownloading CacheStatus = "downloading"
	CacheStatusComplete    CacheStatus = "complete"
	CacheStatusError       CacheStatus = "error"
)

// Valid returns true if the cache status is valid
func (cs CacheStatus) Valid() bool {
	switch cs {
	case CacheStatusPending, CacheStatusDownloading, CacheStatusComplete, CacheStatusError:
		return true
	}
	return false
}

// String returns the string representation
func (cs CacheStatus) String() string {
	return string(cs)
}

// Scan implements the sql.Scanner interface
func (cs *CacheStatus) Scan(value interface{}) error {
	if value == nil {
		*cs = CacheStatusPending
		return nil
	}
	if str, ok := value.(string); ok {
		*cs = CacheStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into CacheStatus", value)
}

// Value implements the driver.Valuer interface
func (cs CacheStatus) Value() (driver.Value, error) {
	return string(cs), nil
}

// RouteOptions represents options for route calculation
type RouteOptions struct {
	RouteType RouteType   `json:"route_type"`
	Vehicle   VehicleType `json:"vehicle"`
	Optimize  bool        `json:"optimize"`
	AvoidTolls bool       `json:"avoid_tolls"`
	AvoidHighways bool    `json:"avoid_highways"`
}

// CreateRouteRequest represents a request to create a route
type CreateRouteRequest struct {
	Name        string        `json:"name" validate:"required,min=1,max=255"`
	Description string        `json:"description"`
	Waypoints   []Point       `json:"waypoints" validate:"required,min=2"`
	Options     RouteOptions  `json:"options"`
}

// CreateGeofenceRequest represents a request to create a geofence
type CreateGeofenceRequest struct {
	Name         string          `json:"name" validate:"required,min=1,max=255"`
	Description  string          `json:"description"`
	Type         GeofenceType    `json:"type" validate:"required"`
	Geometry     interface{}     `json:"geometry" validate:"required"`
	AlertOnEnter bool            `json:"alert_on_enter"`
	AlertOnExit  bool            `json:"alert_on_exit"`
	Enabled      bool            `json:"enabled"`
}

// CreateOfflineAreaRequest represents a request to create an offline area
type CreateOfflineAreaRequest struct {
	Name    string      `json:"name" validate:"required,min=1,max=255"`
	Bounds  BoundingBox `json:"bounds" validate:"required"`
	MinZoom int         `json:"min_zoom" validate:"required,min=1,max=20"`
	MaxZoom int         `json:"max_zoom" validate:"required,min=1,max=20"`
	Layers  []string    `json:"layers" validate:"required"`
}
