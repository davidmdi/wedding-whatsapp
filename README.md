# Wedding WhatsApp RSVP Bot

A Go application that sends wedding invitations via WhatsApp and automatically collects RSVP responses from guests.

## Features

- 📱 Send wedding invitations via WhatsApp
- ✅ Automatic RSVP response handling (YES/NO)
- 📊 Track guest attendance status
- 💾 Persistent storage using JSON files
- 🎨 Interactive CLI for managing guests

## Prerequisites

- Go 1.21 or higher
- WhatsApp account (the bot will connect using QR code)
- **C Compiler** (required for SQLite3):
  - **Windows**: Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MinGW-w64](https://www.mingw-w64.org/)
  - **macOS**: Install Xcode Command Line Tools: `xcode-select --install`
  - **Linux**: Install `gcc` via your package manager (e.g., `sudo apt-get install gcc`)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd wedding-whatsapp
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
# Make sure CGO is enabled (it is by default)
go build -o whatsapp-bot ./cmd/whatsapp-bot

# On Windows, you may need to ensure gcc is in your PATH
# After installing TDM-GCC or MinGW-w64, add it to your PATH
```

**Note**: If you get "CGO_ENABLED=0" errors, ensure you have a C compiler installed and CGO is enabled.

## Configuration

The application uses environment variables for configuration. You can set them or use the defaults:

- `WHATSAPP_DATA_DIR` - Directory for storing WhatsApp session data (default: `data`)
- `WEDDING_DATE` - Date of the wedding (default: `Saturday, January 1, 2025`)
- `WEDDING_LOCATION` - Venue location (default: `Venue TBD`)
- `BRIDE_NAME` - Name of the bride (default: `Bride`)
- `GROOM_NAME` - Name of the groom (default: `Groom`)

### Example Configuration

```bash
export WEDDING_DATE="Saturday, June 15, 2025"
export WEDDING_LOCATION="Grand Ballroom, Hotel XYZ"
export BRIDE_NAME="Sarah"
export GROOM_NAME="John"
export WHATSAPP_DATA_DIR="./data"
```

## Usage

1. Run the application:
```bash
./whatsapp-bot
```

2. On first run, you'll see a QR code. Scan it with WhatsApp:
   - Open WhatsApp on your phone
   - Go to Settings > Linked Devices
   - Tap "Link a Device"
   - Scan the QR code displayed in the terminal

3. Once connected, you can use the interactive CLI:
   - **Option 1**: Send invitation - Enter guest name and phone number to send an invitation
   - **Option 2**: View all guests - See a list of all guests and their RSVP status
   - **Option 3**: View guests by status - Filter guests by pending/accepted/declined
   - **Option 4**: Exit - Close the application

## How It Works

1. **Sending Invitations**: When you send an invitation, the bot creates a guest record and sends a formatted WhatsApp message with wedding details.

2. **RSVP Responses**: Guests can reply with:
   - ✅ **YES** (or variations like "accept", "coming", "will be there")
   - ❌ **NO** (or variations like "decline", "can't come", "won't come")

3. **Automatic Processing**: The bot automatically:
   - Recognizes RSVP responses
   - Updates guest status
   - Sends confirmation messages

## Phone Number Format

When entering phone numbers, use the international format without the `+` sign:
- US: `1234567890`
- UK: `441234567890`
- Example: `1234567890` (not `+1234567890`)

## Data Storage

- Guest data is stored in `{WHATSAPP_DATA_DIR}/guests.json`
- WhatsApp session data is stored in `{WHATSAPP_DATA_DIR}/whatsmeow.db`

## Project Structure

```
wedding-whatsapp/
├── cmd/
│   └── whatsapp-bot/
│       └── main.go          # Main application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── handler/
│   │   └── rsvp.go          # RSVP message handling
│   ├── models/
│   │   └── guest.go         # Guest data model
│   ├── storage/
│   │   └── storage.go       # JSON file storage
│   └── whatsapp/
│       ├── service.go       # WhatsApp service
│       └── logger_adapter.go # Logger adapter
├── go.mod
└── README.md
```

## Troubleshooting

### QR Code Not Appearing
- Make sure you have a stable internet connection
- Try restarting the application

### Messages Not Sending
- Verify the phone number format (country code + number, no spaces or +)
- Ensure the recipient has WhatsApp installed
- Check that you're connected to WhatsApp (look for "Connected to WhatsApp!" message)

### RSVP Not Being Recognized
- Make sure guests reply with clear YES/NO responses
- Check that the guest was added to the system before they reply

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

