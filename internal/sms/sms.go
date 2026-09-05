package sms

type Sender interface {
	SendOTPSMS(phone, code string) error
}