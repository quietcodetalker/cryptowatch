package telegram

type Account struct {
	ID        int64  `json:"id"`
	AuthToken string `json:"auth_token"`
	UserID    uint64 `json:"user_id"`
}

type getUpdatesResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int      `json:"update_id"`
		Message  *Message `json:"message"`
	} `json:"result"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Chat      *Chat  `json:"chat"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
}

type Chat struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}
