package main

import (
	"encoding/xml"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// 文档地址：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Receiving_event_pushes.html

// event type
const (
	subscribe   = "subscribe"
	unsubscribe = "unsubscribe"
	scan        = "SCAN"
	location    = "LOCATION"
	click       = "CLICK"
	view        = "VIEW"
)

type Event struct {
	WeChatCommon
	Event     string // 事件类型，subscribe(订阅)、unsubscribe(取消订阅)，SCAN（扫描已关注的公众号），LOCATION（上报地理位置），CLICK（点击菜单拉取消息），VIEW（点击菜单跳转链接时的）
	EventKey  string // 事件KEY值
	Ticket    string // 二维码的ticket，可用来换取二维码图片
	Latitude  string // 地理位置纬度
	Longitude string // 地理位置经度
	Precision string // 地理位置精度
}

func handlerEvent(body []byte) (interface{}, error) {
	req := &Event{}
	if err := xml.Unmarshal(body, req); err != nil {
		log.Errorf("xml.Unmarshal event request from wechat error: %s", err.Error())
		errors.Wrap(err, "xml.Unmarshal event request from wechat error")
		return nil, err
	}
	switch req.Event {
	case subscribe:
		response := &WeChatMsg{}
		response.FromUserName = req.ToUserName
		response.ToUserName = req.FromUserName
		response.CreateTime = time.Now().Unix()
		response.MsgType = "text"
		response.Content = viper.GetString("event.subscribe")
		return response, nil
	case unsubscribe:
		response := &WeChatMsg{}
		response.FromUserName = req.ToUserName
		response.ToUserName = req.FromUserName
		response.CreateTime = time.Now().Unix()
		response.MsgType = "text"
		response.Content = viper.GetString("event.unsubscribe")
		return response, nil
	case scan:
	case location:
	case click:
	case view:
	default:
		log.Errorf("unknow event: %s", req.Event)
		return nil, errors.Errorf("unknow event: %s", req.Event)

	}
	return nil, nil
}
