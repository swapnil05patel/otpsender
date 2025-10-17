# Email OTP Setup

## For Render Deployment (Recommended)

### Step 1: Setup SendGrid
1. Sign up at https://sendgrid.com (free tier: 100 emails/day)
2. Go to Settings → API Keys
3. Create API Key with "Mail Send" permissions
4. Copy the API key

### Step 2: Configure Render Environment
In Render dashboard, set environment variables:
```
GMAIL_EMAIL=your-verified-sender@domain.com
SENDGRID_API_KEY=your-sendgrid-api-key
BASE_URL=https://your-app.onrender.com
```

## For Local Development

### Step 1: Enable Gmail App Password
1. Go to Google Account settings: https://myaccount.google.com/
2. Security → 2-Step Verification (enable if not enabled)
3. App passwords → Generate app password
4. Select "Mail" and your device
5. Copy the 16-digit app password

### Step 2: Configure Local Credentials
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