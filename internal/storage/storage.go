package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"wedding-whatsapp/internal/models"
)

type Storage struct {
	mu     sync.RWMutex
	guests []models.Guest
	file   string
}

// NewStorage creates a new storage instance
func NewStorage(filePath string) (*Storage, error) {
	s := &Storage{
		guests: make([]models.Guest, 0),
		file:   filePath,
	}

	// Load existing data if file exists
	if _, err := os.Stat(filePath); err == nil {
		if err := s.Load(); err != nil {
			return nil, fmt.Errorf("failed to load storage: %w", err)
		}
	}

	return s, nil
}

// AddGuest adds a new guest or updates existing one
func (s *Storage) AddGuest(guest models.Guest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if guest already exists
	for i, g := range s.guests {
		if g.PhoneNumber == guest.PhoneNumber {
			// Update existing guest
			guest.InvitedDate = g.InvitedDate
			if guest.RSVPStatus == models.RSVPNotInvited {
				guest.RSVPStatus = g.RSVPStatus
			}
			s.guests[i] = guest
			return s.Save()
		}
	}

	// Add new guest
	if guest.InvitedDate.IsZero() {
		guest.InvitedDate = time.Now()
	}
	if guest.RSVPStatus == "" {
		guest.RSVPStatus = models.RSVPPending
	}
	s.guests = append(s.guests, guest)
	return s.Save()
}

// GetGuest retrieves a guest by phone number
func (s *Storage) GetGuest(phoneNumber string) (*models.Guest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, g := range s.guests {
		if g.PhoneNumber == phoneNumber {
			return &g, nil
		}
	}
	return nil, fmt.Errorf("guest not found")
}

// UpdateRSVP updates the RSVP status for a guest
func (s *Storage) UpdateRSVP(phoneNumber string, status models.RSVPStatus, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, g := range s.guests {
		if g.PhoneNumber == phoneNumber {
			s.guests[i].RSVPStatus = status
			s.guests[i].RSVPDate = time.Now()
			if notes != "" {
				s.guests[i].Notes = notes
			}
			return s.Save()
		}
	}
	return fmt.Errorf("guest not found")
}

// GetAllGuests returns all guests
func (s *Storage) GetAllGuests() []models.Guest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guests := make([]models.Guest, len(s.guests))
	copy(guests, s.guests)
	return guests
}

// GetGuestsByStatus returns guests filtered by RSVP status
func (s *Storage) GetGuestsByStatus(status models.RSVPStatus) []models.Guest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Guest
	for _, g := range s.guests {
		if g.RSVPStatus == status {
			result = append(result, g)
		}
	}
	return result
}

// Save saves the guests to file
func (s *Storage) Save() error {
	data, err := json.MarshalIndent(s.guests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(s.file, data, 0644)
}

// Load loads guests from file
func (s *Storage) Load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		s.guests = make([]models.Guest, 0)
		return nil
	}

	if err := json.Unmarshal(data, &s.guests); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}
