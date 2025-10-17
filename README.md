# OTP Sender

Two-step OTP verification system using email and SMS.

## Features
- Email verification with mobile form link
- SMS OTP generation and verification
- Gmail SMTP integration
- Web-based interface

## Setup

1. Configure Gmail credentials in `.env`:
```
GMAIL_EMAIL=your-email@gmail.com
GMAIL_APP_PASSWORD=your-16-digit-app-password
```

2. Run the application:
```
go run gmail-otp.go
```

3. Access: http://localhost:8083

## Flow
1. User enters email → receives mobile form link
2. User enters mobile → receives SMS OTP (console)
3. User verifies OTP → authentication complete

## API Endpoints
- `POST /request-mobile` - Send mobile form link
- `POST /send-sms-otp` - Send OTP to mobile
- `POST /verify-sms-otp` - Verify OTP