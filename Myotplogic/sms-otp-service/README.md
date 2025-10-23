# SMS OTP Service

Free SMS OTP service using TextBelt API.

## Features
- Send OTP to mobile numbers
- Verify OTP codes
- Free SMS service (1 SMS per day per IP)

## Usage

1. Run the server:
   ```
   go run main.go
   ```

2. Send OTP:
   ```
   POST http://localhost:8081/send-otp
   {"phone": "+1234567890"}
   ```

3. Verify OTP:
   ```
   POST http://localhost:8081/verify-otp
   {"phone": "+1234567890", "otp": "123456"}
   ```

## Note
- Uses TextBelt free tier (1 SMS/day per IP)
- For production, use paid SMS services
- Phone numbers must include country code (+1 for US)