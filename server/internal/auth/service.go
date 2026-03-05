package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mathrand "math/rand"

	"github.com/resend/resend-go/v3"
)

func genOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	otp := fmt.Sprintf("%06d", n.Int64())

	return otp, nil
}

func sendOTP(client *resend.Client, email string, otp string) error {
	from := []string{"Khiem", "Anh"}[mathrand.Intn(2)]
	content := fmt.Sprintf(`
		<p>Hello, this is %s from KatanaID</p>
		<div style="font-family: sans-serif; max-width: 400px; margin: 0 auto; padding: 20px;">
		<p>Your verification code is</p>
  	<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; padding: 20px; background: #f4f4f4; text-align: center; border-radius: 8px;">
    %s
  	</div>
  	<p style="color: #666; margin-top: 16px;">It will expire in 5 minutes.</p>
  	<p style="color: #999; font-size: 12px;">If you didn't request this, you can ignore this email.</p>
		</div>
		<p>Thanks, from KatanaID team</p>
	`, from, otp)

	_, err := client.Emails.Send(&resend.SendEmailRequest{
		From:    fmt.Sprintf("%s@katanaid.com", from),
		To:      []string{email},
		Subject: "Your OTP Code for KatanaID",
		Html:    content,
	})

	return err
}
