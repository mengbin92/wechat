package main

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// The ChatRequest struct defines the structure for requests to be sent to the chat service. It contains two fields:
// - Content: the content of the message
// - Tokens: the number of tokens to generate in the response.
type ChatRequest struct {
	Content string `json:"content" form:"content"`
	Tokens  int    `json:"tokens,omitempty" form:"tokens,omitempty"`
}

// The WeChatInfo struct contains three fields:
// - Token: The token provided by WeChat to identify the server.
// - AppID: The ID of the application registered with WeChat
// - Appsecret: The app secret provided by WeChat.
// - EncodingAESKey: The base64 encode aes key
type WeChatInfo struct {
	Token          string
	AppID          string
	Appsecret      string
	EncodingAESKey string
	Signature      WeChatVerify
}

// The WeChatVerify struct is used to verify requests received from WeChat. It contains four fields:
// - Signature: The signature to verify. It is created by hashing a string consisting of the timestamp, nonce and token parameters together with SHA-1 algorithm.
// - Timestamp: A timestamp used to ensure that this request is not replayed.
// - Nonce: A random value used to create a unique hash.
// - Echostr: A parameter used to respond to the verification request.
type WeChatVerify struct {
	Signature    string `json:"signature" form:"signature"`
	Timestamp    string `json:"timestamp" form:"timestamp"`
	Nonce        string `json:"nonce" form:"nonce"`
	Echostr      string `json:"echostr" form:"echostr"`
	EncryptType  string `json:"encrypt_type" form:"encrypt_type,omitempty"`
	MsgSignature string `json:"msg_signature" form:"msg_signature,omitempty"`
}

// The WeChatMsg struct defines the structure for a WeChat message. It contains six fields:
// - XMLName: This is the name of the XML element.
// - ToUserName: The account ID of the recipient
// - FromUserName: The account ID of the sender
// - CreateTime: Time value in Unix format when message sent by sender.
// - MsgType: The type of message. (text, image, voice etc.)
// - Content: The message content
type WeChatMsg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	Content      string
}

// The WeChatEncrypt struct defines the sturcture for a encrypted WeChat message. It has three fields:
// - XMLName: The name of the XML element.
// - ToUserName: The account ID of the recipient.
// - Encrypt: The encrypted message data.
type WeChatEncrypt struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string
	Encrypt    string
}

type EncryptRequest struct {
	AppId string
}

// The Verify method verifies a WeChat request signature.
// It calculates the hash of the token, timestamp and nonce, combines them into a single string and then applies SHA-1 algorithm to generate a hexadecimal-encoded signature. If this generated signature matches with the provided signature then the method returns true.
func (p *WeChatVerify) Verify(token string) bool {
	// var s []string
	// if p.EncryptType == "" {
	s := []string{token, p.Timestamp, p.Nonce}
	// } else {
	// 	s = []string{token, p.Timestamp, p.Nonce, p.MsgSignature}
	// }

	sort.Strings(s)
	str := strings.Join(s, "")
	hashs := sha1.New()
	hashs.Write([]byte(str))

	signature := hex.EncodeToString(hashs.Sum(nil))
	if signature == p.Signature {
		return true
	} else {
		return false
	}
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

func (info *WeChatInfo) EncryptMsg(body []byte, toUser string) (resp []byte, err error) {
	key, err := decodeAESKey(info.EncodingAESKey)
	if err != nil {
		return nil, errors.Wrap(err, "decode AESKey error")
	}
	respBody, err := aesEncrypt(body, key)
	if err != nil {
		return nil, errors.Wrap(err, "aes encrypt response error")
	}
	response := &WeChatEncrypt{
		ToUserName: toUser,
		Encrypt:    string(respBody),
	}
	resp, err = xml.Marshal(response)
	if err != nil {
		return nil, errors.Wrap(err, "marshal encrypt response error")
	}
	return
}

func (info *WeChatInfo) DecryptMsg(body []byte) (random, rawXMLMsg, appID []byte, err error) {
	context := &WeChatEncrypt{}
	err = xml.Unmarshal(body, context)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "unmarshal WeChatEncrypt body error")
	}
	key, err := decodeAESKey(info.EncodingAESKey)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "decode AESKey error")
	}

	ciphertext, _ := base64.StdEncoding.DecodeString(context.Encrypt)
	plainText, err := aesDecrypt(ciphertext, key)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "aes decrypt requset error")
	}

	rawXMLMsgLen := int(decodeNetworkByteOrder(plainText[16:20]))
	if rawXMLMsgLen < 0 {
		errors.Wrapf(err, "incorrect msg length: %d", rawXMLMsgLen)
		return
	}
	appIDOffset := 20 + rawXMLMsgLen
	if len(plainText) <= appIDOffset {
		errors.Wrapf(err, "msg length too large: %d", rawXMLMsgLen)
		return
	}

	random = plainText[:16:20]
	rawXMLMsg = plainText[20:appIDOffset:appIDOffset]
	appID = plainText[appIDOffset:]
	return
}

func (info *WeChatInfo) Verify(encrypt string) bool {
	s := []string{info.Token, info.Signature.Timestamp, info.Signature.Nonce, encrypt}
	sort.Strings(s)
	str := strings.Join(s, "")
	hashs := sha1.New()
	hashs.Write([]byte(str))

	signature := hex.EncodeToString(hashs.Sum(nil))
	if signature == info.Signature.MsgSignature {
		return true
	} else {
		return false
	}
}
