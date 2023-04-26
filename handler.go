package main

import (
	"encoding/base64"
	"encoding/xml"

	"github.com/pkg/errors"
)

func handlerEncrypt(body []byte, timestamp, nonce, msg_sig string) (random, rawXMLMsg []byte, err error) {
	request := &WeChatEncryptRequest{}
	err = xml.Unmarshal(body, request)
	if err != nil {
		log.Errorf("unmarshal wechat encrypt request error: %s")
		errors.Wrap(err, "unmarshal wechat encrypt request error")
		return
	}

	// verify msg from wechat signature
	if calcSignature(token, timestamp, nonce, request.Encrypt) != msg_sig {
		log.Errorf("encrypt msg got from wechat verify signature failed")
		errors.New("encrypt msg got from wechat verify signature failed")
		return
	}

	// decode cipher text from base64
	cipherText, err := base64.StdEncoding.DecodeString(request.Encrypt)
	if err != nil {
		log.Errorf("decode wechat encrypt request error: %s", err.Error())
		errors.Wrap(err, "decode wechat encrypt request error")
		return
	}
	// aes decrypt
	plainText, err := aesDecrypt(cipherText, key)
	if err != nil {
		log.Errorf("decrypt wechat encrypt request error: %s", err.Error())
		errors.Wrap(err, "decrypt wechat encrypt request error")
		return
	}

	// get raw wechat encrypt request length
	rawXMLMsgLen := int(ntohl(plainText[16:20]))
	if rawXMLMsgLen < 0 {
		log.Errorf("incorrect msg length: %d", rawXMLMsgLen)
		errors.Wrapf(err, "incorrect msg length: %d", rawXMLMsgLen)
		return
	}

	// verify appid
	appIDOffset := 20 + rawXMLMsgLen
	if len(plainText) <= appIDOffset {
		log.Errorf("msg length too large: %d", rawXMLMsgLen)
		errors.Wrapf(err, "msg length too large: %d", rawXMLMsgLen)
		return
	}
	// verify appid
	if appID != string(plainText[appIDOffset:]) {
		log.Errorf("Received an attack disguised as a WeChat server.")
		errors.New("Received an attack disguised as a WeChat server.")
		return
	}

	// get random which from wechat
	random = plainText[:16:20]

	// raw wechat msg
	rawXMLMsg = plainText[20:appIDOffset:appIDOffset]
	return
}

func handlerEncryptResponse(random []byte, timestamp, nonce string, msg any) (resp WeChatEncryptResponse, err error) {
	var rawXMLMsg, cipherText []byte
	rawXMLMsg, err = xml.Marshal(msg)
	if err != nil {
		log.Errorf("marshal response msg error: %s", err.Error())
		errors.Wrap(err, "marshal response msg error")
		return
	}

	appIDOffset := 20 + len(rawXMLMsg)
	plainText := make([]byte, appIDOffset+len(appID))

	// create plainText
	copy(plainText[:16], random)
	htonl(plainText[16:20], uint32(len(rawXMLMsg)))
	copy(plainText[20:], rawXMLMsg)
	copy(plainText[appIDOffset:], []byte(appID))

	cipherText, err = aesEncrypt(plainText, key)
	if err != nil {
		log.Errorf("encrypt response error: %s", err.Error())
		errors.Wrap(err, "encrypt response error")
		return
	}

	resp.Encrypt = base64.RawStdEncoding.EncodeToString(cipherText)
	resp.Nonce = nonce
	resp.Timestamp = timestamp
	resp.MsgSignature = calcSignature(token, timestamp, nonce, resp.Encrypt)
	return
}
