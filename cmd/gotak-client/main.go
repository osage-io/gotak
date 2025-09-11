package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dfedick/gotak/pkg/cot"
)

func main() {
	var (
		server   = flag.String("server", "localhost:8087", "TAK server address")
		callsign = flag.String("callsign", "TestClient", "Client callsign")
		protocol = flag.String("protocol", "tcp", "Protocol to use (tcp, udp)")
		lat      = flag.Float64("lat", 37.7749, "Initial latitude")
		lon      = flag.Float64("lon", -122.4194, "Initial longitude")
	)
	flag.Parse()

	fmt.Printf("GoTAK Test Client\n")
	fmt.Printf("Connecting to %s using %s protocol\n", *server, *protocol)
	fmt.Printf("Callsign: %s\n", *callsign)
	fmt.Printf("Position: %.6f, %.6f\n", *lat, *lon)
	fmt.Println()

	switch *protocol {
	case "tcp":
		runTCPClient(*server, *callsign, *lat, *lon)
	case "udp":
		runUDPClient(*server, *callsign, *lat, *lon)
	default:
		log.Fatalf("Unsupported protocol: %s", *protocol)
	}
}

func runTCPClient(server, callsign string, lat, lon float64) {
	// Connect to server
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", server)

	// Start receiving messages
	go receiveMessages(conn)

	// Send initial position report
	sendPositionReport(conn, callsign, lat, lon)

	// Interactive client
	fmt.Println("Commands:")
	fmt.Println("  pos <lat> <lon> - Send position update")
	fmt.Println("  chat <message>  - Send chat message")
	fmt.Println("  ping            - Send ping")
	fmt.Println("  quit            - Exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		command := strings.ToLower(parts[0])

		switch command {
		case "pos":
			if len(parts) < 2 {
				fmt.Println("Usage: pos <lat> <lon>")
				continue
			}
			coords := strings.Fields(parts[1])
			if len(coords) != 2 {
				fmt.Println("Usage: pos <lat> <lon>")
				continue
			}
			var newLat, newLon float64
			if _, err := fmt.Sscanf(coords[0], "%f", &newLat); err != nil {
				fmt.Printf("Invalid latitude: %v\n", err)
				continue
			}
			if _, err := fmt.Sscanf(coords[1], "%f", &newLon); err != nil {
				fmt.Printf("Invalid longitude: %v\n", err)
				continue
			}
			sendPositionReport(conn, callsign, newLat, newLon)

		case "chat":
			if len(parts) < 2 {
				fmt.Println("Usage: chat <message>")
				continue
			}
			sendChatMessage(conn, callsign, parts[1])

		case "ping":
			sendPing(conn)

		case "quit", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

func runUDPClient(server, callsign string, lat, lon float64) {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	// Connect to server
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s (UDP)\n", server)

	// Send initial position report
	sendPositionReport(conn, callsign, lat, lon)

	// Send periodic position updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	fmt.Println("Sending periodic position updates every 30 seconds...")
	fmt.Println("Press Ctrl+C to stop")

	for {
		select {
		case <-ticker.C:
			// Add some small random movement
			newLat := lat + (float64(time.Now().Unix()%10)-5)*0.001
			newLon := lon + (float64(time.Now().Unix()%10)-5)*0.001
			sendPositionReport(conn, callsign, newLat, newLon)
		}
	}
}

func receiveMessages(conn net.Conn) {
	buffer := make([]byte, 8192)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Error reading from server: %v\n", err)
			}
			return
		}

		// Try to parse as CoT message
		event, err := cot.ParseCoT(buffer[:n])
		if err != nil {
			fmt.Printf("Received non-CoT message: %s\n", string(buffer[:n]))
			continue
		}

		// Handle different message types
		switch {
		case cot.IsTypeChat(event.Type):
			if event.Detail != nil && event.Detail.Remarks != nil {
				fmt.Printf("Chat from %s: %s\n", event.Detail.Remarks.Source, event.Detail.Remarks.Text)
			}
		case cot.IsTypeAtom(event.Type) || cot.IsTypeBit(event.Type):
			callsign := event.GetCallsign()
			if lat, lon, err := event.GetPosition(); err == nil {
				fmt.Printf("Position from %s: %.6f, %.6f\n", callsign, lat, lon)
			}
		case cot.IsTypeSystem(event.Type):
			fmt.Printf("System message: %s\n", event.Type)
		default:
			fmt.Printf("Received message type: %s\n", event.Type)
		}
	}
}

func sendPositionReport(conn net.Conn, callsign string, lat, lon float64) {
	uid := fmt.Sprintf("%s-%d", callsign, time.Now().Unix())
	event := cot.NewPositionEvent(uid, callsign, lat, lon, 0)
	
	xmlData, err := event.ToXML()
	if err != nil {
		fmt.Printf("Error creating position report: %v\n", err)
		return
	}

	_, err = conn.Write(xmlData)
	if err != nil {
		fmt.Printf("Error sending position report: %v\n", err)
		return
	}

	fmt.Printf("Sent position: %.6f, %.6f\n", lat, lon)
}

func sendChatMessage(conn net.Conn, callsign, message string) {
	uid := fmt.Sprintf("chat-%s-%d", callsign, time.Now().Unix())
	event := cot.NewChatEvent(uid, callsign, "All", message)
	
	xmlData, err := event.ToXML()
	if err != nil {
		fmt.Printf("Error creating chat message: %v\n", err)
		return
	}

	_, err = conn.Write(xmlData)
	if err != nil {
		fmt.Printf("Error sending chat message: %v\n", err)
		return
	}

	fmt.Printf("Sent chat: %s\n", message)
}

func sendPing(conn net.Conn) {
	event := cot.NewEvent("ping", cot.TypeSystemPing)
	
	xmlData, err := event.ToXML()
	if err != nil {
		fmt.Printf("Error creating ping: %v\n", err)
		return
	}

	_, err = conn.Write(xmlData)
	if err != nil {
		fmt.Printf("Error sending ping: %v\n", err)
		return
	}

	fmt.Println("Sent ping")
}
