package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

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
	reqBody := &WeChatMsg{}
	err = xml.Unmarshal(reqBodyBytes, reqBody)
	if err != nil {
		log.Errorf("xml.Unmarshal request error: %s", err.Error())
		errResponse(ctx, errors.Wrap(err, "xml.Unmarshal request error"))
		return
	}

	reqCache := &WeChatCache{
		OpenID:  reqBody.FromUserName,
		Content: reqBody.Content,
	}

	respBody := &WeChatMsg{}
	respBody.FromUserName = reqBody.ToUserName
	respBody.ToUserName = reqBody.FromUserName
	respBody.CreateTime = time.Now().Unix()
	respBody.MsgType = reqBody.MsgType

	respChan := make(chan string)
	errChan := make(chan error)

	switch reqBody.MsgType {
	case "text":
		reply, err := cache.Get(context.Background(), reqCache.Key()).Bytes()
		if err != nil && len(reply) == 0 {
			log.Info("get nothing from local cache,now get data from openai")

			go goChatWithChan(reqCache, respChan, errChan)

			select {
			case respBody.Content = <-respChan:
			case err := <-errChan:
				respBody.Content = err.Error()
			case <-time.After(4900 * time.Millisecond):
				respBody.Content = "前方网络拥堵....\n等待是为了更好的相遇，稍后请重新发送上面的问题来获取答案，感谢理解"
			}
		} else {
			respBody.Content = string(reply)
		}
	case "event":
	case "image":
	case "voice":
	case "video":
	case "shortvideo":
	case "location":
	case "link":
		log.Infof("MsgType: %s not implemented", reqBody.MsgType)
		errResponse(ctx, fmt.Errorf("MsgType: %s not implemented", reqBody.MsgType))
		return
	default:
		log.Errorf("unknow MsgType: %s", reqBody.MsgType)
		errResponse(ctx, fmt.Errorf("unknow MsgType: %s", reqBody.MsgType))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknow MsgType: %s", reqBody.MsgType)})
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
	// 	after := WeChatEncryptRequest{
	// 		Encrypt:    resp.Encrypt,
	// 		ToUserName: respBody.FromUserName,
	// 	}
	// 	xmlResponse(ctx, after)

	// }
}
