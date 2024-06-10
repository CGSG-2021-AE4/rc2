package api

import "encoding/json"

type readMsgWrapper[ContentT any] struct {
	MsgType string   `json:"msg_type"`
	Content ContentT `json:"content"`
}

type writeMsgWrapper[ContentT any] struct {
	MsgType string   `json:"msg_type"`
	Content ContentT `json:"content"`
}

func ReadMsg[ContentT any](buf []byte) (string, ContentT, error) {
	var msg readMsgWrapper[ContentT]

	if err := json.Unmarshal(buf, &msg); err != nil {
		return "", msg.Content, err
	}
	return msg.MsgType, msg.Content, nil
}

func WriteMsg[ContentT any](mt string, c ContentT) ([]byte, error) {
	return json.Marshal(writeMsgWrapper[ContentT]{
		MsgType: mt,
		Content: c,
	})
}

func WriteError(err string) ([]byte, error) {
	return WriteMsg[string]("error", err)
}
