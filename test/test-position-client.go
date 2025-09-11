package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/dfedick/gotak/pkg/cot"
)

func main() {
	// Connect to the GoTAK server
	conn, err := net.Dial("tcp", "localhost:8087")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to GoTAK server")

	// Send friendly units
	friendlyUnits := []struct {
		callsign string
		lat      float64
		lng      float64
	}{
		{"ALPHA-1", 38.9072, -77.0369}, // Washington DC
		{"BRAVO-2", 38.9100, -77.0400}, // Near Washington DC
		{"CHARLIE-3", 38.9050, -77.0350}, // Near Washington DC
	}

	// Send hostile units
	hostileUnits := []struct {
		callsign string
		lat      float64
		lng      float64
	}{
		{"ENEMY-1", 38.8900, -77.0500}, // South of DC
		{"ENEMY-2", 38.8950, -77.0450}, // South of DC
	}

	// Send initial positions for friendly units
	for i, unit := range friendlyUnits {
		event := cot.NewPositionEvent(
			fmt.Sprintf("friendly-unit-%d", i+1),
			unit.callsign,
			unit.lat,
			unit.lng,
			100.0, // altitude in meters
		)
		
		// Set as friendly (atom)
		event.Type = cot.TypeAtomFriendlyGround
		
		// Add group information
		event.SetGroup("FRIENDLY", "Team Member")
		
		// Send the position report
		xmlData, err := event.ToXML()
		if err != nil {
			log.Printf("Error creating XML for %s: %v", unit.callsign, err)
			continue
		}
		
		_, err = conn.Write(xmlData)
		if err != nil {
			log.Printf("Error sending position for %s: %v", unit.callsign, err)
			continue
		}
		
		log.Printf("Sent initial position for %s at %.6f, %.6f", unit.callsign, unit.lat, unit.lng)
		time.Sleep(500 * time.Millisecond)
	}

	// Send initial positions for hostile units
	for i, unit := range hostileUnits {
		event := cot.NewPositionEvent(
			fmt.Sprintf("hostile-unit-%d", i+1),
			unit.callsign,
			unit.lat,
			unit.lng,
			100.0,
		)
		
		// Set as hostile (bit)
		event.Type = cot.TypeBitHostileGround
		
		// Add group information
		event.SetGroup("HOSTILE", "Enemy Force")
		
		// Send the position report
		xmlData, err := event.ToXML()
		if err != nil {
			log.Printf("Error creating XML for %s: %v", unit.callsign, err)
			continue
		}
		
		_, err = conn.Write(xmlData)
		if err != nil {
			log.Printf("Error sending position for %s: %v", unit.callsign, err)
			continue
		}
		
		log.Printf("Sent initial position for %s at %.6f, %.6f", unit.callsign, unit.lat, unit.lng)
		time.Sleep(500 * time.Millisecond)
	}

	// Now send periodic updates to simulate movement
	log.Println("Starting position updates...")
	
	updateCount := 0
	for {
		// Update friendly units (move in circle pattern)
		for i, unit := range friendlyUnits {
			// Create circular movement
			angle := float64(updateCount) * 0.1 * (float64(i+1)) // Different speed for each unit
			radius := 0.002 // ~200 meters
			
			newLat := unit.lat + radius*math.Cos(angle)
			newLng := unit.lng + radius*math.Sin(angle)
			
			event := cot.NewPositionEvent(
				fmt.Sprintf("friendly-unit-%d", i+1),
				unit.callsign,
				newLat,
				newLng,
				100.0,
			)
			event.Type = cot.TypeAtomFriendlyGround
			event.SetGroup("FRIENDLY", "Team Member")
			
			// Add track data (speed and course)
			if event.Detail == nil {
				event.Detail = &cot.Detail{}
			}
			event.Detail.Track = &cot.Track{
				Speed:  fmt.Sprintf("%.1f", 5.0+float64(i)), // 5-7 m/s
				Course: fmt.Sprintf("%.0f", math.Mod(angle*180/math.Pi, 360)),
			}
			
			xmlData, err := event.ToXML()
			if err != nil {
				log.Printf("Error creating update XML for %s: %v", unit.callsign, err)
				continue
			}
			
			_, err = conn.Write(xmlData)
			if err != nil {
				log.Printf("Error sending update for %s: %v", unit.callsign, err)
				continue
			}
			
			time.Sleep(200 * time.Millisecond)
		}
		
		// Update hostile units (move in straight lines)
		for i, unit := range hostileUnits {
			// Move north slowly
			newLat := unit.lat + float64(updateCount)*0.0001*(float64(i+1)*0.5)
			newLng := unit.lng + float64(updateCount)*0.0001*(float64(i+1)*0.3)
			
			event := cot.NewPositionEvent(
				fmt.Sprintf("hostile-unit-%d", i+1),
				unit.callsign,
				newLat,
				newLng,
				100.0,
			)
			event.Type = cot.TypeBitHostileGround
			event.SetGroup("HOSTILE", "Enemy Force")
			
			// Add track data
			if event.Detail == nil {
				event.Detail = &cot.Detail{}
			}
			event.Detail.Track = &cot.Track{
				Speed:  fmt.Sprintf("%.1f", 3.0+float64(i)*0.5), // 3-3.5 m/s
				Course: "45", // Northeast
			}
			
			xmlData, err := event.ToXML()
			if err != nil {
				log.Printf("Error creating update XML for %s: %v", unit.callsign, err)
				continue
			}
			
			_, err = conn.Write(xmlData)
			if err != nil {
				log.Printf("Error sending update for %s: %v", unit.callsign, err)
				continue
			}
			
			time.Sleep(200 * time.Millisecond)
		}
		
		updateCount++
		log.Printf("Sent update round %d", updateCount)
		
		// Send updates every 5 seconds
		time.Sleep(5 * time.Second)
		
		// Stop after 20 updates (100 seconds)
		if updateCount > 20 {
			break
		}
	}
	
	log.Println("Test complete")
}
