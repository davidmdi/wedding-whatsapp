package handler

import (
	"fmt"
	"strings"

	"wedding-whatsapp/internal/models"
	"wedding-whatsapp/internal/storage"
	"wedding-whatsapp/internal/whatsapp"

	"go.mau.fi/whatsmeow/types/events"
)

type RSVPHandler struct {
	whatsappService *whatsapp.Service
	storage         *storage.Storage
	config          *Config
}

type Config struct {
	WeddingDate     string
	WeddingLocation string
	BrideName       string
	GroomName       string
}

// NewRSVPHandler creates a new RSVP handler
func NewRSVPHandler(whatsappService *whatsapp.Service, storage *storage.Storage, cfg *Config) *RSVPHandler {
	return &RSVPHandler{
		whatsappService: whatsappService,
		storage:         storage,
		config:          cfg,
	}
}

// HandleMessage processes incoming WhatsApp messages for RSVP responses
func (h *RSVPHandler) HandleMessage(msg *events.Message) error {
	if msg.Message == nil {
		return nil
	}

	text := msg.Message.GetConversation()
	if text == "" {
		return nil
	}

	// Get sender phone number
	sender := msg.Info.Sender.String()
	phoneNumber := strings.Split(sender, "@")[0]

	// Normalize phone number (remove + and spaces)
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")

	// Get guest - only process RSVP if guest was previously invited
	_, err := h.storage.GetGuest(phoneNumber)
	if err != nil {
		// Guest not found, might be a new conversation - ignore
		return nil
	}

	// Check if this is an RSVP response
	text = strings.ToLower(strings.TrimSpace(text))

	var newStatus models.RSVPStatus
	var responseMessage string

	if containsAny(text, "yes", "yep", "yeah", "accept", "accepting", "attending", "coming", "will come", "will be there", "✅") {
		newStatus = models.RSVPAccepted
		responseMessage = fmt.Sprintf(
			"🎉 Wonderful! We're so excited to celebrate with you!\n\n"+
				"We've confirmed your attendance for the wedding of %s & %s on %s.\n\n"+
				"See you there! 💕",
			h.config.BrideName, h.config.GroomName, h.config.WeddingDate,
		)
	} else if containsAny(text, "no", "nope", "decline", "declining", "not coming", "can't come", "won't come", "can't make it", "❌") {
		newStatus = models.RSVPDeclined
		responseMessage = fmt.Sprintf(
			"Thank you for letting us know. We're sorry you won't be able to join us for the wedding of %s & %s.\n\n"+
				"We'll miss you! 💕",
			h.config.BrideName, h.config.GroomName,
		)
	} else {
		// Not a clear RSVP response, ignore
		return nil
	}

	// Update RSVP status
	if err := h.storage.UpdateRSVP(phoneNumber, newStatus, ""); err != nil {
		return fmt.Errorf("failed to update RSVP: %w", err)
	}

	// Send confirmation message
	if err := h.whatsappService.SendMessage(phoneNumber, responseMessage); err != nil {
		return fmt.Errorf("failed to send confirmation: %w", err)
	}

	return nil
}

// SendInvitation sends a wedding invitation to a guest
func (h *RSVPHandler) SendInvitation(phoneNumber, name string) error {
	// Normalize phone number before storing (so it matches WhatsApp format)
	normalizedNumber := whatsapp.NormalizePhoneNumber(phoneNumber)

	// Add or update guest in storage with normalized phone number
	guest := models.Guest{
		PhoneNumber: normalizedNumber,
		Name:        name,
		RSVPStatus:  models.RSVPPending,
	}

	if err := h.storage.AddGuest(guest); err != nil {
		return fmt.Errorf("failed to add guest: %w", err)
	}

	// Send invitation via WhatsApp (it will normalize again, but that's fine)
	if err := h.whatsappService.SendInvitation(
		phoneNumber,
		name,
		h.config.WeddingDate,
		h.config.WeddingLocation,
		h.config.BrideName,
		h.config.GroomName,
	); err != nil {
		return fmt.Errorf("failed to send invitation: %w", err)
	}

	return nil
}

// containsAny checks if the text contains any of the given keywords
func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
