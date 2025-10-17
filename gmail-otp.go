package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EmailOTPRequest struct {
	Email string `json:"email"`
}

type MobileRequest struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

type SMSOTPVerifyRequest struct {
	Mobile string `json:"mobile"`
	OTP    string `json:"otp"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var emailVerified = make(map[string]bool)
var mobileOTPStore = make(map[string]string)

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func getBaseURL() string {
	if url := os.Getenv("BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8083"
}

func loadConfig() (string, string) {
	// Try environment variables first (for Render)
	if email := os.Getenv("GMAIL_EMAIL"); email != "" {
		return email, os.Getenv("GMAIL_APP_PASSWORD")
	}

	// Fallback to .env file for local development
	file, err := os.Open(".env")
	if err != nil {
		return "", ""
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			config[parts[0]] = parts[1]
		}
	}
	return config["GMAIL_EMAIL"], config["GMAIL_APP_PASSWORD"]
}

func sendEmail(to, subject, body, from, password string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body)

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "smtp.gmail.com",
	}

	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", tlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}

func requestMobile(w http.ResponseWriter, r *http.Request) {
	var req EmailOTPRequest
	json.NewDecoder(r.Body).Decode(&req)

	from, password := loadConfig()
	subject := "Mobile Number Required for OTP"
	body := fmt.Sprintf("Please provide your mobile number to receive OTP.\n\nClick: %s/mobile-form?email=%s", getBaseURL(), req.Email)

	err := sendEmail(req.Email, subject, body, from, password)
	if err != nil {
		fmt.Printf("Email Error: %v\n", err)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Failed to send email"})
		return
	}

	emailVerified[req.Email] = true
	fmt.Printf("Mobile request sent to %s\n", req.Email)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "Check email for mobile form link"})
}

func sendSMSOTP(w http.ResponseWriter, r *http.Request) {
	var req MobileRequest
	json.NewDecoder(r.Body).Decode(&req)

	if !emailVerified[req.Email] {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Email not verified"})
		return
	}

	otp := generateOTP()
	mobileOTPStore[req.Mobile] = otp

	// Simulate SMS sending
	fmt.Printf("SMS to %s: Your OTP is %s\n", req.Mobile, otp)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "OTP sent to mobile (check console)"})
}

func verifySMSOTP(w http.ResponseWriter, r *http.Request) {
	var req SMSOTPVerifyRequest
	json.NewDecoder(r.Body).Decode(&req)

	if storedOTP, exists := mobileOTPStore[req.Mobile]; exists && storedOTP == req.OTP {
		delete(mobileOTPStore, req.Mobile)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "Mobile OTP verified successfully"})
	} else {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid OTP"})
	}
}

func mobileForm(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>Mobile Number</title></head>
<body>
<h2>Enter Mobile Number</h2>
<p>Email: %s</p>
<input type="text" id="mobile" placeholder="+1234567890">
<button onclick="submitMobile()">Send OTP to Mobile</button>
<div id="result"></div>
<script>
function submitMobile() {
	const mobile = document.getElementById('mobile').value;
	fetch('/send-sms-otp', {
		method: 'POST',
		headers: {'Content-Type': 'application/json'},
		body: JSON.stringify({email: '%s', mobile: mobile})
	}).then(r => r.json()).then(data => {
		document.getElementById('result').innerHTML = JSON.stringify(data);
		if(data.success) {
			window.location.href = '/verify-form?mobile=' + mobile;
		}
	});
}
</script>
</body>
</html>`, email, email)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func verifyForm(w http.ResponseWriter, r *http.Request) {
	mobile := r.URL.Query().Get("mobile")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>Verify OTP</title></head>
<body>
<h2>Verify OTP</h2>
<p>Mobile: %s</p>
<input type="text" id="otp" placeholder="123456">
<button onclick="verifyOTP()">Verify OTP</button>
<div id="result"></div>
<script>
function verifyOTP() {
	const otp = document.getElementById('otp').value;
	fetch('/verify-sms-otp', {
		method: 'POST',
		headers: {'Content-Type': 'application/json'},
		body: JSON.stringify({mobile: '%s', otp: otp})
	}).then(r => r.json()).then(data => {
		document.getElementById('result').innerHTML = JSON.stringify(data);
	});
}
</script>
</body>
</html>`, mobile, mobile)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/request-mobile", requestMobile)
	http.HandleFunc("/send-sms-otp", sendSMSOTP)
	http.HandleFunc("/verify-sms-otp", verifySMSOTP)
	http.HandleFunc("/mobile-form", mobileForm)
	http.HandleFunc("/verify-form", verifyForm)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("Gmail OTP Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}