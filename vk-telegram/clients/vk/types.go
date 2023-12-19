package vk

type Connection struct {
	Response Response `json:"response"`
}

type Response struct {
	Key    string `json:"key"`
	Server string `json:"server"`
	Ts     string `json:"ts"`
}

type ResultTsUpdates struct {
	Ts      string
	Updates []Update
}

type Update struct {
	Type   string `json:"type"`
	Object Object `json:"object"`
}

type Object struct {
	Message Message `json:"message"`
}

type Message struct {
	ChatID      int          `json:"from_id"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Type  string `json:"type,omitempty"`
	Photo Photo  `json:"photo,omitempty"`
}

type Photo struct {
	Sizes []Size `json:"sizes,omitempty"`
}

type Size struct {
	URL string `json:"url,omitempty"`
}
