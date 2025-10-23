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
	"sync"
	"time"
)

type EmailOTPRequest struct {
	Email string `json:"email"`
}

type EmailOTPVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type OTPData struct {
	OTP       string
	ExpiresAt time.Time
}

var (
	emailOTPStore = make(map[string]OTPData)
	mutex         = sync.RWMutex{}
)

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func loadConfig() (string, string) {
	file, err := os.Open(".env")
	if err != nil {
		fmt.Println("Enter Gmail credentials:")
		fmt.Print("Gmail Email: ")
		reader := bufio.NewReader(os.Stdin)
		email, _ := reader.ReadString('\n')
		fmt.Print("Gmail App Password: ")
		password, _ := reader.ReadString('\n')
		return strings.TrimSpace(email), strings.TrimSpace(password)
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

func sendEmailOTP(w http.ResponseWriter, r *http.Request) {
	var req EmailOTPRequest
	json.NewDecoder(r.Body).Decode(&req)

	otp := generateOTP()
	expiresAt := time.Now().Add(2 * time.Minute)

	mutex.Lock()
	emailOTPStore[req.Email] = OTPData{OTP: otp, ExpiresAt: expiresAt}
	mutex.Unlock()

	// Send email concurrently
	go func() {
		from, password := loadConfig()
		subject := "Your OTP Code"
		body := fmt.Sprintf("Your OTP is: %s\nValid for 2 minutes.", otp)

		err := sendEmail(req.Email, subject, body, from, password)
		if err != nil {
			fmt.Printf("Email Error: %v\n", err)
			fmt.Printf("OTP for %s: %s\n", req.Email, otp)
		} else {
			fmt.Printf("Email sent to %s with OTP: %s\n", req.Email, otp)
		}
	}()

	fmt.Printf("OTP generated for %s: %s (expires in 2 minutes)\n", req.Email, otp)
	json.NewEncoder(w).Encode(Response{Success: true, Message: "OTP sent to email"})
}

func verifyEmailOTP(w http.ResponseWriter, r *http.Request) {
	var req EmailOTPVerifyRequest
	json.NewDecoder(r.Body).Decode(&req)

	mutex.Lock()
	defer mutex.Unlock()

	otpData, exists := emailOTPStore[req.Email]
	if !exists {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "No OTP found for this email"})
		return
	}

	if time.Now().After(otpData.ExpiresAt) {
		delete(emailOTPStore, req.Email)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "OTP has expired"})
		return
	}

	if otpData.OTP == req.OTP {
		delete(emailOTPStore, req.Email)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "OTP verified successfully"})
	} else {
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid OTP"})
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/send-email-otp", sendEmailOTP)
	http.HandleFunc("/verify-email-otp", verifyEmailOTP)
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "email-test.html")
		} else {
			http.NotFound(w, r)
		}
	})

	log.Println("Gmail OTP Server starting on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}