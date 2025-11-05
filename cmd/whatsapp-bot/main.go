package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"wedding-whatsapp/internal/config"
	"wedding-whatsapp/internal/handler"
	"wedding-whatsapp/internal/models"
	"wedding-whatsapp/internal/storage"
	"wedding-whatsapp/internal/whatsapp"
)

func main() {
	fmt.Println("🎉 Wedding WhatsApp RSVP Bot")
	fmt.Println("============================")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize storage
	storagePath := fmt.Sprintf("%s/guests.json", cfg.WhatsAppDataDir)
	guestStorage, err := storage.NewStorage(storagePath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Initialize WhatsApp service
	whatsappCfg := &whatsapp.Config{
		DataDir: cfg.WhatsAppDataDir,
	}
	whatsappService, err := whatsapp.NewService(whatsappCfg)
	if err != nil {
		fmt.Printf("Error initializing WhatsApp service: %v\n", err)
		os.Exit(1)
	}

	// Initialize RSVP handler
	rsvpHandler := handler.NewRSVPHandler(whatsappService, guestStorage, &handler.Config{
		WeddingDate:     "05.01.2026",
		WeddingLocation: "אולמי אמרה נס ציונה",
		BrideName:       "ענת מגן",
		GroomName:       "דוד מדינרדזה",
	})

	// Set message handler
	whatsappService.SetMessageHandler(rsvpHandler.HandleMessage)

	// Connect to WhatsApp
	fmt.Println("Connecting to WhatsApp...")
	if err := whatsappService.Connect(); err != nil {
		fmt.Printf("Error connecting to WhatsApp: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Connected to WhatsApp!")
	fmt.Println("The bot is now listening for RSVP responses.\n")

	// Start interactive CLI
	go startCLI(rsvpHandler, guestStorage, cfg)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n\nShutting down...")
	whatsappService.Disconnect()
	fmt.Println("Goodbye! 👋")
}

func startCLI(rsvpHandler *handler.RSVPHandler, storage *storage.Storage, cfg *config.Config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nCommands:")
		fmt.Println("  1. Send invitation")
		fmt.Println("  2. View all guests")
		fmt.Println("  3. View guests by status")
		fmt.Println("  4. Exit")
		fmt.Print("\nEnter command (1-4): ")

		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())

		switch command {
		case "1":
			sendInvitation(scanner, rsvpHandler)
		case "2":
			viewAllGuests(storage)
		case "3":
			viewGuestsByStatus(scanner, storage)
		case "4":
			fmt.Println("Exiting...")
			os.Exit(0)
		default:
			fmt.Println("Invalid command. Please try again.")
		}
	}
}

func sendInvitation(scanner *bufio.Scanner, rsvpHandler *handler.RSVPHandler) {
	fmt.Print("Enter guest name: ")
	if !scanner.Scan() {
		return
	}
	name := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter phone number (with country code, e.g., 1234567890): ")
	if !scanner.Scan() {
		return
	}
	phoneNumber := strings.TrimSpace(scanner.Text())

	// Normalize phone number
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")

	fmt.Printf("\nSending invitation to %s (%s)...\n", name, phoneNumber)
	if err := rsvpHandler.SendInvitation(phoneNumber, name); err != nil {
		fmt.Printf("❌ Error sending invitation: %v\n", err)
	} else {
		fmt.Printf("✅ Invitation sent successfully!\n")
	}
}

func viewAllGuests(storage *storage.Storage) {
	guests := storage.GetAllGuests()
	if len(guests) == 0 {
		fmt.Println("\nNo guests found.")
		return
	}

	fmt.Printf("\n📋 All Guests (%d total):\n", len(guests))
	fmt.Println(strings.Repeat("-", 60))
	for _, guest := range guests {
		fmt.Printf("Name: %s\n", guest.Name)
		fmt.Printf("Phone: %s\n", guest.PhoneNumber)
		fmt.Printf("Status: %s\n", guest.RSVPStatus)
		if !guest.RSVPDate.IsZero() {
			fmt.Printf("RSVP Date: %s\n", guest.RSVPDate.Format("2006-01-02 15:04:05"))
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}

func viewGuestsByStatus(scanner *bufio.Scanner, storage *storage.Storage) {
	fmt.Println("\nSelect status:")
	fmt.Println("  1. Pending")
	fmt.Println("  2. Accepted")
	fmt.Println("  3. Declined")
	fmt.Print("Enter choice (1-3): ")

	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	var status models.RSVPStatus

	switch choice {
	case "1":
		status = models.RSVPPending
	case "2":
		status = models.RSVPAccepted
	case "3":
		status = models.RSVPDeclined
	default:
		fmt.Println("Invalid choice.")
		return
	}

	guests := storage.GetGuestsByStatus(status)
	if len(guests) == 0 {
		fmt.Printf("\nNo guests with status '%s'.\n", string(status))
		return
	}

	fmt.Printf("\n📋 Guests with status '%s' (%d total):\n", string(status), len(guests))
	fmt.Println(strings.Repeat("-", 60))
	for _, guest := range guests {
		fmt.Printf("Name: %s\n", guest.Name)
		fmt.Printf("Phone: %s\n", guest.PhoneNumber)
		if !guest.RSVPDate.IsZero() {
			fmt.Printf("RSVP Date: %s\n", guest.RSVPDate.Format("2006-01-02 15:04:05"))
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}
