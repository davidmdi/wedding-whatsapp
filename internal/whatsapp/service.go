package whatsapp

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// messageHandler is a callback function for handling messages
type MessageHandler func(*events.Message) error

type Config struct {
	DataDir string
}

type Service struct {
	client         *whatsmeow.Client
	cfg            *Config
	log            zerolog.Logger
	messageHandler MessageHandler
}

// NewService creates a new WhatsApp service
func NewService(cfg *Config) (*Service, error) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Str("component", "WhatsApp").Logger()

	// Use nil logger - sqlstore will use a no-op logger by default
	container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s/whatsmeow.db?_foreign_keys=on", cfg.DataDir), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Use nil logger - whatsmeow will use a no-op logger by default
	client := whatsmeow.NewClient(deviceStore, nil)

	service := &Service{
		client: client,
		cfg:    cfg,
		log:    logger,
	}

	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		service.eventHandler(evt)
	})

	return service, nil
}

// NormalizePhoneNumber normalizes phone numbers to international format
// Handles Israeli numbers that start with 0 by converting to +972 format
func NormalizePhoneNumber(phoneNumber string) string {
	// Remove all non-digit characters
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")

	// Handle Israeli phone numbers (starting with 0)
	// Israeli format: 05XXXXXXXX -> 9725XXXXXXXX
	if strings.HasPrefix(phoneNumber, "0") && len(phoneNumber) == 10 {
		// Remove leading 0 and add country code 972
		phoneNumber = "972" + phoneNumber[1:]
	}

	// Handle Israeli numbers already with country code but wrong format
	// If it starts with 9720, remove the 0 after 972
	if strings.HasPrefix(phoneNumber, "9720") {
		phoneNumber = "972" + phoneNumber[4:]
	}

	return phoneNumber
}

// Connect connects to WhatsApp
func (s *Service) Connect() error {
	if s.client.Store.ID == nil {
		qrChan, _ := s.client.GetQRChannel(context.Background())
		err := s.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Generate and display QR code in terminal
				q, err := qrcode.New(evt.Code, qrcode.Medium)
				if err != nil {
					fmt.Printf("QR Code: %s\n", evt.Code)
					fmt.Println("Please scan this QR code with WhatsApp to connect.")
				} else {
					fmt.Println("\n" + q.ToSmallString(false))
					fmt.Println("📱 Please scan the QR code above with WhatsApp:")
					fmt.Println("   1. Open WhatsApp on your phone")
					fmt.Println("   2. Go to Settings > Linked Devices")
					fmt.Println("   3. Tap 'Link a Device'")
					fmt.Println("   4. Scan the QR code shown above\n")
				}
			} else {
				fmt.Printf("Login event: %s\n", evt.Event)
			}
		}
	} else {
		err := s.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}
	return nil
}

// Disconnect disconnects from WhatsApp
func (s *Service) Disconnect() {
	s.client.Disconnect()
}

// SendInvitation sends a wedding invitation with RSVP buttons
func (s *Service) SendInvitation(phoneNumber, name, weddingDate, weddingLocation, brideName, groomName string) error {
	message := fmt.Sprintf(
		"🎉 *Wedding Invitation*\n\n"+
			"Dear %s,\n\n"+
			"You are cordially invited to celebrate the wedding of\n\n"+
			"*%s* & *%s*\n\n"+
			"📅 Date: %s\n"+
			"📍 Location: %s\n\n"+
			"Please confirm your attendance by selecting one of the options below.",
		name, brideName, groomName, weddingDate, weddingLocation,
	)

	// Normalize phone number before parsing
	phoneNumber = NormalizePhoneNumber(phoneNumber)

	// Create JID - try with + prefix first (WhatsApp sometimes prefers this format)
	var jid types.JID
	var err error

	// Try with + prefix
	if parsedJID, parseErr := types.ParseJID("+" + phoneNumber); parseErr == nil {
		jid = parsedJID
	} else {
		// Try without + prefix
		if parsedJID, parseErr := types.ParseJID(phoneNumber); parseErr == nil {
			jid = parsedJID
		} else {
			// Create JID with phone number and default user server (without +)
			jid = types.NewJID(phoneNumber, types.DefaultUserServer)
		}
	}

	// Verify the number is on WhatsApp before sending
	resp, verifyErr := s.client.IsOnWhatsApp(context.Background(), []string{phoneNumber})
	if verifyErr != nil {
		return fmt.Errorf("failed to verify number on WhatsApp: %w", verifyErr)
	}

	if len(resp) == 0 || !resp[0].IsIn {
		return fmt.Errorf("number %s is not registered on WhatsApp or not in contacts. Please ensure: 1) The number has WhatsApp, 2) The number is saved in your phone contacts with country code (e.g., +972...), 3) WhatsApp has synced contacts", phoneNumber)
	}

	// Use the verified JID from WhatsApp
	jid = resp[0].JID

	// Log verification result
	fmt.Printf("✓ Number verified on WhatsApp: %s (JID: %s)\n", phoneNumber, jid.String())

	// For now, we'll send a simple message with text instructions
	// as interactive buttons require specific WhatsApp Business API setup
	message += "\n\nReply with:\n✅ *YES* to accept\n❌ *NO* to decline"

	// Log the JID being used for debugging
	s.log.Debug().Str("jid", jid.String()).Str("phone", phoneNumber).Msg("Attempting to send message")

	sentMsg, err := s.client.SendMessage(context.Background(), jid, &waE2E.Message{
		Conversation: &message,
	})

	if err == nil {
		fmt.Printf("✓ Message sent successfully! ID: %s, Timestamp: %v\n", sentMsg.ID, sentMsg.Timestamp)
	}

	if err != nil {
		// Provide more helpful error message
		if strings.Contains(err.Error(), "unknown server") || strings.Contains(err.Error(), "can't send message") {
			return fmt.Errorf("failed to send message to %s (JID: %s): %w. Note: The recipient must be in your WhatsApp contacts. Try: 1) Ensure the number is in your phone contacts with country code (972...), 2) Wait for WhatsApp to sync contacts (may take a few minutes), 3) Or have them message you first", phoneNumber, jid.String(), err)
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendMessage sends a simple text message
func (s *Service) SendMessage(phoneNumber, message string) error {
	// Normalize phone number before parsing
	phoneNumber = NormalizePhoneNumber(phoneNumber)

	// Create JID - try with + prefix first (WhatsApp sometimes prefers this format)
	var jid types.JID
	var err error

	// Try with + prefix
	if parsedJID, parseErr := types.ParseJID("+" + phoneNumber); parseErr == nil {
		jid = parsedJID
	} else {
		// Try without + prefix
		if parsedJID, parseErr := types.ParseJID(phoneNumber); parseErr == nil {
			jid = parsedJID
		} else {
			// Create JID with phone number and default user server (without +)
			jid = types.NewJID(phoneNumber, types.DefaultUserServer)
		}
	}

	// Verify the number is on WhatsApp before sending
	resp, verifyErr := s.client.IsOnWhatsApp(context.Background(), []string{phoneNumber})
	if verifyErr != nil {
		return fmt.Errorf("failed to verify number on WhatsApp: %w", verifyErr)
	}

	if len(resp) == 0 || !resp[0].IsIn {
		return fmt.Errorf("number %s is not registered on WhatsApp or not in contacts. Please ensure: 1) The number has WhatsApp, 2) The number is saved in your phone contacts with country code (e.g., +972...), 3) WhatsApp has synced contacts", phoneNumber)
	}

	// Use the verified JID from WhatsApp
	jid = resp[0].JID

	// Log verification result
	fmt.Printf("✓ Number verified on WhatsApp: %s (JID: %s)\n", phoneNumber, jid.String())

	// Log the JID being used for debugging
	s.log.Debug().Str("jid", jid.String()).Str("phone", phoneNumber).Msg("Attempting to send message")

	sentMsg, err := s.client.SendMessage(context.Background(), jid, &waE2E.Message{
		Conversation: &message,
	})

	if err == nil {
		fmt.Printf("✓ Message sent successfully! ID: %s, Timestamp: %v\n", sentMsg.ID, sentMsg.Timestamp)
	}

	if err != nil {
		// Provide more helpful error message
		if strings.Contains(err.Error(), "unknown server") || strings.Contains(err.Error(), "can't send message") {
			return fmt.Errorf("failed to send message to %s (JID: %s): %w. Note: The recipient must be in your WhatsApp contacts. Try: 1) Ensure the number is in your phone contacts with country code (972...), 2) Wait for WhatsApp to sync contacts (may take a few minutes), 3) Or have them message you first", phoneNumber, jid.String(), err)
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// eventHandler handles incoming WhatsApp events
func (s *Service) eventHandler(evt interface{}) {
	if evt == nil {
		return
	}
	switch evt := evt.(type) {
	case *events.Message:
		s.handleMessage(evt)
	case *events.Connected:
		s.log.Info().Msg("Connected to WhatsApp")
	case *events.Disconnected:
		s.log.Info().Msg("Disconnected from WhatsApp")
	case *events.LoggedOut:
		s.log.Info().Msg("Logged out from WhatsApp")
	}
}

// handleMessage processes incoming messages
func (s *Service) handleMessage(msg *events.Message) {
	// Skip messages from self
	if msg.Info.IsFromMe {
		return
	}

	// Call custom message handler if set
	if s.messageHandler != nil {
		if err := s.messageHandler(msg); err != nil {
			s.log.Error().Err(err).Msg("Error handling message")
		}
	} else {
		s.log.Info().
			Str("sender", msg.Info.Sender.String()).
			Str("message", msg.Message.GetConversation()).
			Msg("Received message")
	}
}

// SetMessageHandler sets a custom handler for incoming messages
func (s *Service) SetMessageHandler(handler MessageHandler) {
	s.messageHandler = handler
}
