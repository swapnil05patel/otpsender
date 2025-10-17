# Gmail OTP Setup

## Step 1: Enable Gmail App Password

1. Go to Google Account settings: https://myaccount.google.com/
2. Security → 2-Step Verification (enable if not enabled)
3. App passwords → Generate app password
4. Select "Mail" and your device
5. Copy the 16-digit app password

## Step 2: Configure Credentials

Edit `.env` file:
```
GMAIL_EMAIL=your-email@gmail.com
GMAIL_APP_PASSWORD=your-16-digit-app-password
```

## Step 3: Run Service

```
go run gmail-otp.go
```

Access: http://localhost:8083

## Test

```
curl -X POST http://localhost:8083/send-email-otp -H "Content-Type: application/json" -d "{\"email\":\"recipient@gmail.com\"}"
```

The OTP will be sent to the recipient's email address.