package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MSG91Sender struct {
	authKey  string
	flowId   string
	senderId string
}

func NewMSG91Sender(authKey, flowId, senderId string) *MSG91Sender {
	return &MSG91Sender{authKey: authKey, flowId: flowId, senderId: senderId}
}

type msg91Recipient struct {
	Mobiles string `json:"mobile"`
	Var1    string `json:"VAR1"`
}

type msg91Request struct {
	FlowId    string           `json:"flow_id"`
	Sender    string           `json:"sender"`
	Recipient []msg91Recipient `json:"recipient"`
}

type msg91Responce struct {
	Massege string `json:"massege"`
	Type    string `json:"type"`
}

func (s MSG91Sender) SendOTPSMS(phone, code string) error {
	payload := msg91Request{
		FlowId: s.flowId,
		Sender: s.senderId,
		Recipient: []msg91Recipient{
			{Mobiles: phone, Var1: code},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("emcode msg91 request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://control.msg91.com/api/v5/flow", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build msg91 err: %w", err)
	}

	req.Header.Set("content-Type", "application/json")
	req.Header.Set("Authkey", s.authKey)

	responce, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send msg91 request: %w", err)
	}
	defer responce.Body.Close()

	var result msg91Responce

	if err := json.NewDecoder(responce.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode msg91 responce err: %w", err)
	}

	if result.Type != "success" {
		return fmt.Errorf("msg91 send failed: %s", result.Massege)
	}

	return nil
}