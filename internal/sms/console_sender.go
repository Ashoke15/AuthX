package sms

type ConsoleSender struct{}

func NewConsolSender() *ConsoleSender {
	return &ConsoleSender{}
}

func (s*ConsoleSender) SendOTPSMS(phone, code string) error {
	println("== OTP SMS ==")
	println("To:", phone)
	println("code:", code)
	println("==============")

	return nil
}