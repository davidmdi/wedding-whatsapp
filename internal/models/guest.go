package models

import "time"

// Guest represents a wedding guest
type Guest struct {
	PhoneNumber string     `json:"phone_number"`
	Name        string     `json:"name"`
	RSVPStatus  RSVPStatus `json:"rsvp_status"`
	RSVPDate    time.Time  `json:"rsvp_date,omitempty"`
	InvitedDate time.Time  `json:"invited_date"`
	Notes       string     `json:"notes,omitempty"`
}

// RSVPStatus represents the attendance confirmation status
type RSVPStatus string

const (
	RSVPPending    RSVPStatus = "pending"
	RSVPAccepted   RSVPStatus = "accepted"
	RSVPDeclined   RSVPStatus = "declined"
	RSVPNotInvited RSVPStatus = "not_invited"
)

// AttendanceRequest represents a request to send an invitation
type AttendanceRequest struct {
	PhoneNumber string
	Name        string
	Message     string
}
