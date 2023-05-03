package main

import (
	"crypto/sha512"
	"encoding/hex"
)

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

type Answer struct {
	Key   string
	Reply string
}

type Answers struct {
	Answers []Answer
}

func (a *Answers) Reply(key string) string {
	for _,answer := range a.Answers{
		if key == answer.Key{
			return answer.Reply
		}
	}
	return ""
}