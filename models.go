package main

import (
	"crypto/sha512"
	"encoding/hex"
)

// The ChatRequest struct defines the structure for requests to be sent to the chat service. It contains two fields:
// - Content: the content of the message
// - Tokens: the number of tokens to generate in the response.
type ChatRequest struct {
	Content string `json:"content" form:"content"`
	Tokens  int    `json:"tokens,omitempty" form:"tokens,omitempty"`
}

// The WeChatCache struct defines the structure to cache WeChat messages.
// It contains two fields:
// - OpenID: The unique identifier of the recipient's account.
// - Content: The content of WeChat message.
type WeChatCache struct {
	OpenID  string `json:"open_id"`
	Content string `json:"content"`
}

// The Key method generates a unique key for caching a WeChat message by combining the OpenID and content properties of the cache object
// and then applying SHA-512/384 algorithm to create a hexadecimal-encoded key.
func (cache *WeChatCache) Key() string {
	hash := sha512.New384()
	hash.Write([]byte(cache.OpenID + "-" + cache.Content))
	return hex.EncodeToString(hash.Sum(nil))
}
