package common

func SanitizePhone(phone string) string {
	if len(phone) == 13 && phone[4] == '9' {
		phone = phone[0:4] + phone[5:]
	}
	return phone
}
