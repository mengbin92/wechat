package main

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func initWeChatInfo() {
	appID = viper.GetString("wechat.appID")
	appsecret = viper.GetString("wechat.appsecret")
	token = viper.GetString("wechat.token")
	encodingAESKey = viper.GetString("wechat.encodingkey")

	var err error
	key, err = decodeAESKey(encodingAESKey)
	if err != nil {
		panic(errors.Wrap(err, "decode aes key error"))
	}
}

func weChatVerify(ctx *gin.Context) {
	log.Info("Get Msg from wechat")
	var (
		signature = ctx.Query("signature")
		timestamp = ctx.Query("timestamp")
		nonce     = ctx.Query("nonce")
		echostr   = ctx.Query("echostr")
	)
	if calcSignature(token, timestamp, nonce) != signature {
		log.Error("WeChat Verify failed")
		errResponse(ctx, errors.New("WeChat Verify failed"))
		return
	}
	stringResponse(ctx, echostr)
}

func weChat(ctx *gin.Context) {
	log.Info("Get Msg from wechat")

	var (
		signature    = ctx.Query("signature")
		timestamp    = ctx.Query("timestamp")
		nonce        = ctx.Query("nonce")
		encrypt_type = ctx.Query("encrypt_type")
		msg_sig      = ctx.Query("msg_signature")
	)
	if calcSignature(token, timestamp, nonce) != signature {
		log.Error("WeChat Verify failed")
		errResponse(ctx, errors.New("WeChat Verify failed"))
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Errorf("read request body error: %s", err.Error())
		errResponse(ctx, errors.New("read request body error"))
		return
	}
	log.Infof("request body: %s", string(body))

	var reqBodyBytes, random []byte
	if encrypt_type == "" {
		reqBodyBytes = body
	} else {
		random, reqBodyBytes, err = handlerEncrypt(body, timestamp, nonce, msg_sig)
		if err != nil {
			errResponse(ctx, err)
			return
		}
	}
	log.Infof("random: %s", string(random))
	log.Infof("request body: %s", string(reqBodyBytes))

	common := &WeChatCommon{}
	err = xml.Unmarshal(reqBodyBytes, common)
	if err != nil {
		log.Errorf("xml.Unmarshal request error: %s", err.Error())
		errResponse(ctx, errors.Wrap(err, "xml.Unmarshal request error"))
		return
	}
	var respBody interface{}

	switch common.MsgType {
	case "text":
		respBody, err = handlerMessage(reqBodyBytes)
		if err != nil {
			log.Errorf("handlerMessage error: %s", err.Error())
			errResponse(ctx, err)
			return
		}
	case "event":
		respBody, err = handlerEvent(reqBodyBytes)
		if err != nil {
			log.Errorf("handlerEvent error: %s", err.Error())
			errResponse(ctx, err)
			return
		}
	case "image":
	case "voice":
	case "video":
	case "shortvideo":
	case "location":
	case "link":
		log.Infof("MsgType: %s not implemented", common.MsgType)
		errResponse(ctx, fmt.Errorf("MsgType: %s not implemented", common.MsgType))
		return
	default:
		log.Errorf("unknow MsgType: %s", common.MsgType)
		errResponse(ctx, fmt.Errorf("unknow MsgType: %s", common.MsgType))
		return
	}
	// if encrypt_type == "" {
	xmlResponse(ctx, respBody)
	// } else {
	// 	resp, err := handlerEncryptResponse(random, timestamp, nonce, respBody)
	// 	if err != nil {
	// 		log.Errorf("handlerEncryptResponse error: %s", err.Error())
	// 		errResponse(ctx, err)
	// 		return
	// 	}
	// 	respBytes, _ := sonic.Marshal(&resp)
	// 	log.Infof("response: %s", string(respBytes))
	// 	xmlResponse(ctx, resp)
	// }
}
