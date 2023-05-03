# Wechat
a Wechat  official account assistant based on OpenAI.  

## 简介

项目地址在[这里](https://github.com/mengbin92/wechat)，使用[OpenAI](https://platform.openai.com/docs/api-reference/chat)提供的API接口接入ChatGPT服务作为微信公众号的应答后台，提供对话服务。

> 应答服务的响应速度除了网络原因外，主要受使用的API Key影响，如果使用的是免费的API Key的话，响应速度会比付费版的要慢很多。

目前只支持文本消息的应答，支持预设简单问答（可通过配置文件新增或删减）。

## 使用

本项目使用支持本地运行和docker部署两种方式，推荐使用docker部署，这样当配置文件变动后可以自动加载，不需要再手动重启服务：  

> 如果使用的是免费的API Key，由于得到OpenAI响应的时间不确定，建议使用缓存，将用户发送过来的请求进行缓存以提高用户再次提问时的响应速度。

```yaml
version: 1.0.0

# 服务使用的端口
port: 9999

# openAI相关配置
openai:
  # openAI提供的APIKey
  apikey: OPENAI_APIKEY
  # 所属组织，非必填
  org:
  # 代理，非必填
  proxy: http://172.26.128.1:1080

# 配置日志输出级别 
log:
  # 支持debug(-1)、info(0)、warn(1)、error(2)、dpanic(3)、panic(4)、fatal(5)
  level: 0

# 使用redis缓存
redis:
  addr: localhost:6379
  db: 0
  password: PASSWORD
  expire: 600

# 微信公众号服务相关配置
wechat:
  # 公众号预留的token，用于校验接收到的请求是否来自微信服务器
  token: chatGPT
  # 微信提供的AppID，暂未使用
  appID: APPID
  # 微信提供的AppSecret，暂未使用
  appsecret: APPSECRET
  # 安全模式使用的密钥
  encodingkey: ENCODINGKEY

# 预设的应答对话
answers:
	# 关键字，匹配不到时调用openAI服务做应答
  - key: 123
  	# 应答
    reply: 345
  - key: 456
    reply: 2345
```

`docker-compose.yaml`文件内容如下：

```yaml
version: '3.3'

services:
  wechat:
    image: mengbin92/wechat:latest
    container_name: wechat
    volumes:
      - ./config/wechat:/app/conf
      - ./wfi.sh:/app/wfi.sh
    command: /app/wfi.sh -d redis:6379 -c '/app/watcher'
    environment:
      - GIN_MODE=debug
    ports:
      - 18080:9999
    depends_on:
      - redis
    networks:
      - wechat

  redis:
    image: redis:7.0.10
    container_name: redis
    volumes:
      - ./data/redis:/data
      - ./config/redis/redis.conf:/usr/local/etc/redis/redis.conf
    ports:
      - 6379:6379
    networks:
      - wechat

networks:
  wechat:
```

```bash
# 启动服务
$ docker-compose up -d
# 关停服务
$ docker-compose down
```
