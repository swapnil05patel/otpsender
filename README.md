# Email OTP Service

A concurrent email OTP verification service built with Go.

## Features

- Email OTP generation and verification
- 2-minute OTP expiry
- Concurrent email sending for multiple users
- Thread-safe operations
- Gmail SMTP integration

## Setup

1. Create a `.env` file:
```
GMAIL_EMAIL=your-email@gmail.com
GMAIL_APP_PASSWORD=your-app-password
```

2. Run the server:
```bash
go run gmail-otp.go
```

3. Access the frontend at `http://localhost:8083`

## API Endpoints

- `POST /send-email-otp` - Send OTP to email
- `POST /verify-email-otp` - Verify OTP
- `GET /` - Frontend interface

## Requirements

- Go 1.16+
- Gmail account with app password enabled
