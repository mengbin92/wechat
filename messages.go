package main

import "encoding/xml"

type WeChatCommon struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   int64
	MsgType      string
	MsgId        string
	MsgDataId    string
	Idx          string
}

// The WeChatMsg struct defines the structure for a WeChat message. It contains six fields:
// - XMLName: This is the name of the XML element.
// - ToUserName: The account ID of the recipient
// - FromUserName: The account ID of the sender
// - CreateTime: Time value in Unix format when message sent by sender.
// - MsgType: The type of message. (text, image, voice etc.)
// - Content: The message content
type WeChatMsg struct {
	WeChatCommon
	Content string
}

// The WeChatEncryptRequest struct defines the sturcture for a encrypted WeChat message. It has three fields:
// - XMLName: The name of the XML element.
// - ToUserName: The account ID of the recipient.
// - Encrypt: The encrypted message data.
type WeChatEncryptRequest struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string
	Encrypt    string
}

type WeChatEncryptResponse struct {
	XMLName      xml.Name `xml:"xml"`
	Encrypt      string
	MsgSignature string
	Timestamp    string
	Nonce        string
}

type WeChatImg struct {
	WeChatCommon
	PicUrl  string
	MediaId string
}

type WeChatVoice struct {
	WeChatCommon
	MediaId string
	Format  string
}

type WeChatVideo struct {
	WeChatCommon
	MediaId      string
	ThumbMediaId string
}

type WeChatShortVideo struct {
	WeChatCommon
	MediaId      string
	ThumbMediaId string
}

type WeChatLocation struct {
	WeChatCommon
	Location_X string
	Location_Y string
	Scale      string
	Label      string
}

type WeChatLink struct {
	WeChatCommon
	Title       string
	Description string
	Url         string
}
