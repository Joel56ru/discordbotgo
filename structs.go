package main

type Refresh struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
	CreatedAt    int64  `json:"created_at,omitempty"`
}
type Topic struct {
	Id         int    `json:"id"`
	TopicTitle string `json:"topic_title"`
	Body       string `json:"body"`
	Forum      Forum  `json:"forum"`
}
type Forum struct {
	Id        int    `json:"id"`
	Position  int    `json:"position"`
	Name      string `json:"name"`
	Permalink string `json:"permalink"`
	Url       string `json:"url"`
}
