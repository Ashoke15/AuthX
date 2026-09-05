package sms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TwilioSender struct {
	accountSID string
	authToken  string
	fromNumber string
}

func NewTwilioSender(accountSID, authToken, fromNumber string) *TwilioSender {
	return &TwilioSender{accountSID: accountSID, authToken: authToken, fromNumber: fromNumber}
}

func (s *TwilioSender) SendOTPSMS(phone, code string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.accountSID)

	from := url.Values{}
	from.Set("To", phone)
	from.Set("From", s.fromNumber)
	from.Set("Body", fmt.Sprintf("your verification code is %s. It expaire in 10 minute", code))

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(from.Encode()))
	if err != nil {
		return fmt.Errorf("build twilio request err: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.accountSID, s.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send twilio request err %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio send failed: status %d", resp.StatusCode)
	}

	return nil
}