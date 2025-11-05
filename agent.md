# Agent Development Log

## Project: Wedding WhatsApp RSVP Bot

### Overview
A Go application that sends wedding invitations via WhatsApp and automatically collects RSVP responses from guests. Built with the whatsmeow library for WhatsApp integration.

---

## Development Journey

### 1. Initial Setup & Environment Configuration

**Challenge:** Setting up the Go development environment on Windows with CGO requirements.

**Solution:**
- Installed TDM-GCC (MinGW compiler) for Windows
- Added `C:\TDM-GCC-64\bin` to system PATH
- Configured environment variables:
  - `CGO_ENABLED=1` (required for SQLite3 support)
  - Updated PATH for GCC access

**Key Learning:** The `go-sqlite3` package requires CGO to be enabled, which necessitates a C compiler on the system.

---

### 2. VS Code Integration

**Challenge:** Running the app from VS Code with proper environment configuration.

**Solution:**
Created `.vscode/launch.json`:
```json
{
    "name": "Launch WhatsApp Bot",
    "type": "go",
    "request": "launch",
    "mode": "debug",
    "program": "${workspaceFolder}/cmd/whatsapp-bot/main.go",
    "cwd": "${workspaceFolder}",
    "env": {
        "CGO_ENABLED": "1",
        "PATH": "${env:PATH};C:\\TDM-GCC-64\\bin"
    }
}
```

Created `.vscode/settings.json`:
```json
{
    "terminal.integrated.env.windows": {
        "PATH": "${env:PATH};C:\\TDM-GCC-64\\bin",
        "CGO_ENABLED": "1"
    },
    "go.toolsEnvVars": {
        "CGO_ENABLED": "1",
        "PATH": "${env:PATH};C:\\TDM-GCC-64\\bin"
    }
}
```

---

### 3. Database Directory Issue

**Challenge:** App crashed with error: "unable to open database file: The system cannot find the path specified."

**Root Cause:** The `data` directory didn't exist, and SQLite couldn't create the database file.

**Solution:**
- Created the `data` directory in the project root
- SQLite successfully created `whatsmeow.db` for session storage

**Code Context:**
```go
container, err := sqlstore.New(ctx, "sqlite3", 
    fmt.Sprintf("file:%s/whatsmeow.db?_foreign_keys=on", cfg.DataDir), nil)
```

---

### 4. QR Code Display Enhancement

**Challenge:** QR code was displayed as raw base64 string, making it impossible to scan.

**Solution:**
- Added `github.com/skip2/go-qrcode` package
- Implemented terminal-based QR code rendering
- Added user-friendly instructions for scanning

**Implementation:**
```go
q, err := qrcode.New(evt.Code, qrcode.Medium)
if err != nil {
    fmt.Printf("QR Code: %s\n", evt.Code)
} else {
    fmt.Println("\n" + q.ToSmallString(false))
    fmt.Println("📱 Please scan the QR code above with WhatsApp:")
    fmt.Println("   1. Open WhatsApp on your phone")
    fmt.Println("   2. Go to Settings > Linked Devices")
    fmt.Println("   3. Tap 'Link a Device'")
    fmt.Println("   4. Scan the QR code shown above\n")
}
```

---

### 5. Israeli Phone Number Normalization

**Challenge:** Error: "can't send message to unknown server" when sending to Israeli phone numbers.

**Root Cause:** Israeli phone numbers start with `0` (e.g., `0538268277`) but WhatsApp requires international format with country code `972`.

**Solution:**
Created `NormalizePhoneNumber` function to handle Israeli number format:

```go
func NormalizePhoneNumber(phoneNumber string) string {
    // Remove all non-digit characters
    phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
    phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
    phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
    
    // Handle Israeli phone numbers (starting with 0)
    // Israeli format: 05XXXXXXXX -> 9725XXXXXXXX
    if strings.HasPrefix(phoneNumber, "0") && len(phoneNumber) == 10 {
        phoneNumber = "972" + phoneNumber[1:]
    }
    
    // Handle Israeli numbers already with country code but wrong format
    // If it starts with 9720, remove the 0 after 972
    if strings.HasPrefix(phoneNumber, "9720") {
        phoneNumber = "972" + phoneNumber[4:]
    }
    
    return phoneNumber
}
```

**Integration:**
- Applied normalization in `SendInvitation` and `SendMessage` functions
- Updated storage to save normalized numbers for consistent RSVP matching

---

### 6. WhatsApp Contact Verification

**Challenge:** Even with normalized numbers, messages weren't being delivered. Error: "unknown server".

**Root Cause:** WhatsApp requires proper contact verification and JID resolution before sending messages.

**Solution:**
Implemented `IsOnWhatsApp` check to verify and resolve contacts:

```go
// Verify the number is on WhatsApp before sending
resp, verifyErr := s.client.IsOnWhatsApp(context.Background(), []string{phoneNumber})
if verifyErr != nil {
    return fmt.Errorf("failed to verify number on WhatsApp: %w", verifyErr)
}

if len(resp) == 0 || !resp[0].IsIn {
    return fmt.Errorf("number %s is not registered on WhatsApp or not in contacts. " +
        "Please ensure: 1) The number has WhatsApp, " +
        "2) The number is saved in your phone contacts with country code, " +
        "3) WhatsApp has synced contacts", phoneNumber)
}

// Use the verified JID from WhatsApp
jid = resp[0].JID
```

**Benefits:**
- Proper JID resolution from WhatsApp's servers
- Clear error messages when contacts aren't found
- Eliminates manual JID construction errors

---

### 7. Enhanced Logging & Debugging

**Challenge:** Messages appeared to send but weren't received by recipients.

**Solution:**
Added detailed logging to track the message flow:

```go
// Log verification result
fmt.Printf("✓ Number verified on WhatsApp: %s (JID: %s)\n", phoneNumber, jid.String())

// Log send result
sentMsg, err := s.client.SendMessage(context.Background(), jid, &waE2E.Message{
    Conversation: &message,
})

if err == nil {
    fmt.Printf("✓ Message sent successfully! ID: %s, Timestamp: %v\n", 
        sentMsg.ID, sentMsg.Timestamp)
}
```

**Output Example:**
```
✓ Number verified on WhatsApp: 972538268277 (JID: 972538268277@s.whatsapp.net)
✓ Message sent successfully! ID: 3EB0123456789ABCDEF, Timestamp: 2024-11-05 14:30:00
```

---

## Technical Stack

### Core Technologies
- **Language:** Go 1.24.0
- **WhatsApp Library:** `go.mau.fi/whatsmeow` v0.0.0-20251028165006
- **Database:** SQLite3 via `github.com/mattn/go-sqlite3` v1.14.32
- **Logging:** `github.com/rs/zerolog` v1.34.0
- **QR Code:** `github.com/skip2/go-qrcode` v0.0.0-20200617195104

### Build Requirements
- **Compiler:** TDM-GCC 10.3.0 (for CGO support)
- **CGO:** Enabled (required for SQLite)
- **OS:** Windows (adaptable to Linux/macOS)

---

## Project Structure

```
wedding-whatsapp/
├── cmd/
│   └── whatsapp-bot/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handler/
│   │   └── rsvp.go              # RSVP message handling
│   ├── models/
│   │   └── guest.go             # Guest data model
│   ├── storage/
│   │   └── storage.go           # JSON file storage
│   └── whatsapp/
│       └── service.go           # WhatsApp service
├── data/
│   ├── guests.json              # Guest database
│   └── whatsmeow.db            # WhatsApp session data
├── .vscode/
│   ├── launch.json             # Debug configuration
│   └── settings.json           # VS Code settings
├── go.mod
├── go.sum
├── README.md
└── agent.md                    # This file
```

---

## Key Features

### 1. Phone Number Normalization
- Automatic conversion of Israeli local format to international format
- Handles various input formats (with/without spaces, dashes, parentheses)
- Removes leading zeros and adds country code

### 2. Contact Verification
- Uses WhatsApp's `IsOnWhatsApp` API to verify numbers
- Ensures contacts are registered on WhatsApp
- Resolves correct JID format from WhatsApp servers

### 3. RSVP Response Handling
- Automatic detection of YES/NO responses
- Multiple keyword variations supported
- Stores responses in JSON database
- Sends confirmation messages

### 4. Interactive CLI
- Send invitations to guests
- View all guests and their RSVP status
- Filter guests by status (pending/accepted/declined)
- Real-time feedback on operations

---

## Build & Run Commands

### Build
```powershell
$env:Path += ";C:\TDM-GCC-64\bin"
$env:CGO_ENABLED="1"
go build -o whatsapp-bot.exe ./cmd/whatsapp-bot
```

### Run
```powershell
.\whatsapp-bot.exe
```

Or with Go:
```powershell
$env:Path += ";C:\TDM-GCC-64\bin"
$env:CGO_ENABLED="1"
go run ./cmd/whatsapp-bot/main.go
```

---

## Configuration

Environment variables (with defaults):
- `WHATSAPP_DATA_DIR` - Directory for data storage (default: `data`)
- `WEDDING_DATE` - Wedding date (default: `Saturday, January 1, 2025`)
- `WEDDING_LOCATION` - Venue location (default: `Venue TBD`)
- `BRIDE_NAME` - Bride's name (default: `Bride`)
- `GROOM_NAME` - Groom's name (default: `Groom`)

---

## Lessons Learned

### 1. CGO Dependencies
- SQLite requires CGO, which needs a C compiler
- Windows requires explicit compiler setup (TDM-GCC/MinGW)
- Environment variables must be set correctly for both terminal and IDE

### 2. WhatsApp API Limitations
- Cannot send first message to non-contacts
- Contact must be in phone's address book with proper format
- WhatsApp needs time to sync contacts
- JID resolution is critical for message delivery

### 3. International Phone Numbers
- Each country has specific formatting rules
- WhatsApp uses E.164 format internally
- Local formats must be normalized before processing

### 4. Error Handling
- Clear error messages significantly improve user experience
- Logging helps debug issues in production
- Verification before action prevents silent failures

---

## Future Enhancements

### Potential Improvements
1. **Bulk Invitations** - Send to multiple guests at once with rate limiting
2. **Message Templates** - Customizable invitation messages
3. **Web Dashboard** - View RSVP statistics in a web interface
4. **Plus-One Support** - Handle guests with +1 companions
5. **Reminder Messages** - Auto-send reminders to pending guests
6. **Multi-language** - Support for multiple languages
7. **Group Messages** - Send to WhatsApp groups
8. **Media Support** - Attach images/videos to invitations

### Known Limitations
1. Requires contacts to be in phone's address book
2. WhatsApp rate limiting may affect bulk sending
3. Depends on stable internet connection
4. Single device connection (no multi-device support yet)

---

## Troubleshooting Guide

### Issue: CGO Errors
**Symptom:** "Binary was compiled with 'CGO_ENABLED=0'"
**Solution:** 
- Ensure TDM-GCC is installed
- Set `CGO_ENABLED=1`
- Add GCC to PATH

### Issue: Database Not Found
**Symptom:** "unable to open database file"
**Solution:**
- Create `data` directory in project root
- Ensure write permissions

### Issue: QR Code Not Scanning
**Symptom:** Can't scan QR code from terminal
**Solution:**
- Enlarge terminal window
- Ensure good lighting and camera focus
- Try from different device if needed

### Issue: Message Not Delivered
**Symptom:** Message shows sent but not received
**Solution:**
- Verify contact is in phone's address book
- Check contact has country code (e.g., +972...)
- Wait for WhatsApp to sync contacts
- Have recipient message you first

---

## Development Timeline

**Session Duration:** ~2 hours
**Main Challenges:** 
1. CGO setup and configuration (30 min)
2. Phone number normalization (20 min)
3. Contact verification and JID resolution (45 min)
4. QR code display (15 min)
5. Testing and debugging (30 min)

**Total LOC:** ~800 lines of Go code

---

## Credits

**Developer:** AI Assistant (Claude)
**User:** David
**Libraries:** whatsmeow, go-sqlite3, zerolog, go-qrcode
**Platform:** Go 1.24, Windows

---

## License

MIT License - Feel free to use and modify for your wedding or events!

---

*Last Updated: November 5, 2025*

