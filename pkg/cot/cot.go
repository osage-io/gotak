package cot

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Event represents a CoT (Cursor on Target) event message
type Event struct {
	XMLName xml.Name `xml:"event"`
	Version string   `xml:"version,attr"`
	UID     string   `xml:"uid,attr"`
	Type    string   `xml:"type,attr"`
	Time    string   `xml:"time,attr"`
	Start   string   `xml:"start,attr"`
	Stale   string   `xml:"stale,attr"`
	How     string   `xml:"how,attr"`
	
	// Child elements
	Point  *Point  `xml:"point,omitempty"`
	Detail *Detail `xml:"detail,omitempty"`
}

// Point represents the geographic location of a CoT event
type Point struct {
	XMLName xml.Name `xml:"point"`
	Lat     string   `xml:"lat,attr"`
	Lon     string   `xml:"lon,attr"`
	Hae     string   `xml:"hae,attr"` // Height above ellipsoid
	CE      string   `xml:"ce,attr"`  // Circular error
	LE      string   `xml:"le,attr"`  // Linear error
}

// Detail contains additional CoT event details
type Detail struct {
	XMLName  xml.Name `xml:"detail"`
	InnerXML string   `xml:",innerxml"`
	
	// Common detail elements
	Contact    *Contact    `xml:"contact,omitempty"`
	Group      *Group      `xml:"__group,omitempty"`
	Status     *Status     `xml:"status,omitempty"`
	Track      *Track      `xml:"track,omitempty"`
	Takv       *Takv       `xml:"takv,omitempty"`
	Remarks    *Remarks    `xml:"remarks,omitempty"`
	Link       *Link       `xml:"link,omitempty"`
	Emergency  *Emergency  `xml:"emergency,omitempty"`
}

// Contact represents contact information
type Contact struct {
	XMLName   xml.Name `xml:"contact"`
	Endpoint  string   `xml:"endpoint,attr"`
	Callsign  string   `xml:"callsign,attr"`
	Name      string   `xml:"name,attr"`
}

// Group represents group membership information
type Group struct {
	XMLName xml.Name `xml:"__group"`
	Name    string   `xml:"name,attr"`
	Role    string   `xml:"role,attr"`
}

// Status represents status information
type Status struct {
	XMLName   xml.Name `xml:"status"`
	Battery   string   `xml:"battery,attr"`
	ReadyToReceive string `xml:"readyToReceive,attr"`
}

// Track represents tracking information
type Track struct {
	XMLName xml.Name `xml:"track"`
	Speed   string   `xml:"speed,attr"`
	Course  string   `xml:"course,attr"`
}

// Takv represents TAK version information
type Takv struct {
	XMLName xml.Name `xml:"takv"`
	Device  string   `xml:"device,attr"`
	Platform string  `xml:"platform,attr"`
	OS      string   `xml:"os,attr"`
	Version string   `xml:"version,attr"`
}

// Remarks contains text remarks
type Remarks struct {
	XMLName xml.Name `xml:"remarks"`
	Text    string   `xml:",chardata"`
	Source  string   `xml:"source,attr"`
	To      string   `xml:"to,attr"`
	Keywords string  `xml:"keywords,attr"`
}

// Link represents a link to another CoT event
type Link struct {
	XMLName  xml.Name `xml:"link"`
	UID      string   `xml:"uid,attr"`
	Type     string   `xml:"type,attr"`
	Relation string   `xml:"relation,attr"`
}

// Emergency represents emergency information
type Emergency struct {
	XMLName xml.Name `xml:"emergency"`
	Type    string   `xml:"type,attr"`
	Cancel  string   `xml:"cancel,attr"`
}

// CoTType constants for common CoT types
const (
	// Atoms (friendly units)
	TypeAtomFriendlyGround   = "a-f-G"
	TypeAtomFriendlyAir      = "a-f-A"
	TypeAtomFriendlyNaval    = "a-f-S"
	
	// Bits (hostile units)
	TypeBitHostileGround     = "a-h-G"
	TypeBitHostileAir        = "a-h-A"
	TypeBitHostileNaval      = "a-h-S"
	
	// Tasking
	TypeTaskingGeoFence      = "b-m-p-s-p-i"
	TypeTaskingRoute         = "b-m-r"
	
	// Chat messages
	TypeChatMessage          = "b-t-f"
	TypeChatEmergency        = "b-t-f-e"
	
	// System messages
	TypeSystemHeartbeat      = "t-x-c-t"
	TypeSystemPing           = "t-x-c-p"
)

// How constants for CoT "how" attribute
const (
	HowMachineGenerated = "m-g"
	HowHumanGenerated   = "h-g"
	HowPredicted        = "h-p"
	HowExtrapolated     = "h-e"
)

// ParseCoT parses a CoT XML message into an Event struct
func ParseCoT(xmlData []byte) (*Event, error) {
	var event Event
	
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("failed to parse CoT XML: %w", err)
	}
	
	// Validate required fields
	if event.UID == "" {
		return nil, fmt.Errorf("CoT event missing required UID")
	}
	if event.Type == "" {
		return nil, fmt.Errorf("CoT event missing required type")
	}
	if event.Time == "" {
		return nil, fmt.Errorf("CoT event missing required time")
	}
	
	return &event, nil
}

// ToXML converts a CoT Event to XML bytes
func (e *Event) ToXML() ([]byte, error) {
	// Set XML namespace if not present
	if e.XMLName.Local == "" {
		e.XMLName = xml.Name{Local: "event"}
	}
	
	// Set version if not present
	if e.Version == "" {
		e.Version = "2.0"
	}
	
	xmlData, err := xml.MarshalIndent(e, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CoT to XML: %w", err)
	}
	
	// Add XML header
	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	return append([]byte(xmlHeader), xmlData...), nil
}

// IsValid checks if the CoT event is valid
func (e *Event) IsValid() error {
	if e.UID == "" {
		return fmt.Errorf("missing UID")
	}
	if e.Type == "" {
		return fmt.Errorf("missing type")
	}
	if e.Time == "" {
		return fmt.Errorf("missing time")
	}
	if e.Start == "" {
		return fmt.Errorf("missing start")
	}
	if e.Stale == "" {
		return fmt.Errorf("missing stale")
	}
	if e.How == "" {
		return fmt.Errorf("missing how")
	}
	
	// Validate time format
	if _, err := time.Parse(time.RFC3339Nano, e.Time); err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}
	
	return nil
}

// IsStale checks if the CoT event is stale
func (e *Event) IsStale() bool {
	staleTime, err := time.Parse(time.RFC3339Nano, e.Stale)
	if err != nil {
		return true // If we can't parse, assume stale
	}
	return time.Now().After(staleTime)
}

// GetPosition returns the lat/lon position if available
func (e *Event) GetPosition() (lat, lon float64, err error) {
	if e.Point == nil {
		return 0, 0, fmt.Errorf("no position data")
	}
	
	lat, err = strconv.ParseFloat(e.Point.Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %w", err)
	}
	
	lon, err = strconv.ParseFloat(e.Point.Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %w", err)
	}
	
	return lat, lon, nil
}

// GetCallsign returns the callsign from contact details
func (e *Event) GetCallsign() string {
	if e.Detail == nil || e.Detail.Contact == nil {
		return ""
	}
	return e.Detail.Contact.Callsign
}

// GetGroup returns the group name
func (e *Event) GetGroup() string {
	if e.Detail == nil || e.Detail.Group == nil {
		return ""
	}
	return e.Detail.Group.Name
}

// SetPosition sets the geographic position
func (e *Event) SetPosition(lat, lon, hae float64) {
	if e.Point == nil {
		e.Point = &Point{}
	}
	e.Point.Lat = fmt.Sprintf("%.8f", lat)
	e.Point.Lon = fmt.Sprintf("%.8f", lon)
	e.Point.Hae = fmt.Sprintf("%.2f", hae)
}

// SetContact sets contact information
func (e *Event) SetContact(callsign, endpoint string) {
	if e.Detail == nil {
		e.Detail = &Detail{}
	}
	if e.Detail.Contact == nil {
		e.Detail.Contact = &Contact{}
	}
	e.Detail.Contact.Callsign = callsign
	e.Detail.Contact.Endpoint = endpoint
}

// SetGroup sets group information
func (e *Event) SetGroup(name, role string) {
	if e.Detail == nil {
		e.Detail = &Detail{}
	}
	if e.Detail.Group == nil {
		e.Detail.Group = &Group{}
	}
	e.Detail.Group.Name = name
	e.Detail.Group.Role = role
}

// NewEvent creates a new CoT event with basic required fields
func NewEvent(uid, eventType string) *Event {
	now := time.Now()
	return &Event{
		XMLName: xml.Name{Local: "event"},
		Version: "2.0",
		UID:     uid,
		Type:    eventType,
		Time:    now.Format(time.RFC3339Nano),
		Start:   now.Format(time.RFC3339Nano),
		Stale:   now.Add(5 * time.Minute).Format(time.RFC3339Nano),
		How:     HowMachineGenerated,
	}
}

// NewPositionEvent creates a new position report event
func NewPositionEvent(uid, callsign string, lat, lon, hae float64) *Event {
	event := NewEvent(uid, TypeAtomFriendlyGround)
	event.SetPosition(lat, lon, hae)
	event.SetContact(callsign, "")
	return event
}

// NewChatEvent creates a new chat message event
func NewChatEvent(uid, from, to, message string) *Event {
	event := NewEvent(uid, TypeChatMessage)
	
	if event.Detail == nil {
		event.Detail = &Detail{}
	}
	
	// Add remarks for chat message
	event.Detail.Remarks = &Remarks{
		Text:   message,
		Source: from,
		To:     to,
	}
	
	return event
}

// IsTypeAtom returns true if the CoT type represents an "atom" (friendly unit)
func IsTypeAtom(cotType string) bool {
	return strings.HasPrefix(cotType, "a-f-")
}

// IsTypeBit returns true if the CoT type represents a "bit" (hostile unit)
func IsTypeBit(cotType string) bool {
	return strings.HasPrefix(cotType, "a-h-")
}

// IsTypeChat returns true if the CoT type represents a chat message
func IsTypeChat(cotType string) bool {
	return strings.HasPrefix(cotType, "b-t-f")
}

// IsTypeSystem returns true if the CoT type represents a system message
func IsTypeSystem(cotType string) bool {
	return strings.HasPrefix(cotType, "t-x-")
}
