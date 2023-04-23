package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func initWeChatInfo() {
	weChatInfo = &WeChatInfo{
		AppID:          viper.GetString("wechat.appID"),
		Appsecret:      viper.GetString("wechat.appsecret"),
		Token:          viper.GetString("wechat.token"),
		EncodingAESKey: viper.GetString("wechat.encodingkey"),
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
	if !verify(weChatInfo.Token, timestamp, nonce, signature) {
		log.Error("WeChat Verify failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "WeChat Verify failed"})
		return
	}
	ctx.Writer.WriteString(echostr)
}

func weChat(ctx *gin.Context) {
	log.Info("Get Msg from wechat")

	var (
		signature    = ctx.Query("signature")
		timestamp    = ctx.Query("timestamp")
		nonce        = ctx.Query("nonce")
		encrypt_type = ctx.Query("encrypt_type")
		// msg_sig      = ctx.Query("msg_signature")
	)

	if !verify(weChatInfo.Token, timestamp, nonce, signature) {
		log.Error("WeChat Verify failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "WeChat Verify failed"})
		return
	}

	log.Info("verify pass")

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Errorf("read request body error: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "read request body error"})
		return
	}
	log.Infof("request body: %s", string(body))

	var reqBodyBytes []byte
	if encrypt_type == "" {
		reqBodyBytes = body
	} else {
		_, reqBodyBytes, _, err = weChatInfo.DecryptMsg(body)
		if err != nil {
			log.Errorf("decrypt msg error: %s", err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "read request body error"})
			return
		}
	}
	log.Infof("request body: %s", string(reqBodyBytes))
	reqBody := &WeChatMsg{}
	err = xml.Unmarshal(reqBodyBytes, reqBody)
	if err != nil {
		log.Errorf("xml.Unmarshal request error: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "read request body error"})
		return
	}
	reqBytes, _ := sonic.Marshal(reqBody)
	log.Infof("Get requset from wechat: %s", string(reqBytes))

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
				// default:
				// 	resp.Content = "答案整理中，请30s稍后重试"
			}
		} else {
			respBody.Content = string(reply)
		}
		respBodyBytes, _ := xml.Marshal(respBody)
		log.Infof("return msg to wechat: %s", string(respBodyBytes))
		var respBytes []byte
		if encrypt_type == "" {
			respBytes = respBodyBytes
		} else {
			respBytes, _ = weChatInfo.EncryptMsg(respBodyBytes, respBody.FromUserName)
		}
		ctx.Writer.Header().Set("Content-Type", "text/xml")
		ctx.Writer.WriteString(string(respBytes))
	default:
		log.Errorf("unknow MsgType: %s", reqBody.MsgType)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknow MsgType: %s", reqBody.MsgType)})
		return
	}
}
