package main

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/pkg/errors"
)

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

func handlerMessage(body []byte) (response *WeChatMsg, err error) {
	reqBody := &WeChatMsg{}
	err = xml.Unmarshal(body, reqBody)
	if err != nil {
		log.Errorf("xml.Unmarshal request error: %s", err.Error())
		errors.Wrap(err, "xml.Unmarshal request error")
		return
	}

	reqCache := &WeChatCache{
		OpenID:  reqBody.FromUserName,
		Content: reqBody.Content,
	}

	response = &WeChatMsg{}
	response.FromUserName = reqBody.ToUserName
	response.ToUserName = reqBody.FromUserName
	response.CreateTime = time.Now().Unix()
	response.MsgType = reqBody.MsgType

	respChan := make(chan string)
	errChan := make(chan error)

	a := answers.Reply(reqBody.Content)
	if a != "" {
		response.Content = a
	} else {
		reply, err := cache.Get(context.Background(), reqCache.Key()).Bytes()
		if err != nil && len(reply) == 0 {
			log.Info("get nothing from local cache,now get data from openai")
			go goChatWithChan(reqCache, respChan, errChan)

			select {
			case response.Content = <-respChan:
			case err := <-errChan:
				response.Content = err.Error()
			case <-time.After(4900 * time.Millisecond):
				response.Content = "前方网络拥堵....\n等待是为了更好的相遇，稍后请重新发送上面的问题来获取答案，感谢理解"
			}
		} else {
			response.Content = string(reply)
		}
	}

	return
}
