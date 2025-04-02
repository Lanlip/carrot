package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	Token        = ""
	GatewayURL   = "wss://gateway.discord.gg/?v=9&encoding=json"
	TargetUserID = "1233392373543993364"
	TargetEmoji  = "🥕"
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:          500,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
		DisableKeepAlives:     false,
	},
	Timeout: 500 * time.Millisecond,
}

func main() {
	for {
		connect()
		time.Sleep(time.Second)
	}
}

func connect() {
	ws, _, err := websocket.DefaultDialer.Dial(GatewayURL, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	if err = ws.WriteJSON(map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token": Token,
			"properties": map[string]string{
				"os": "Windows", "browser": "Chrome", "device": "",
			},
		},
	}); err != nil {
		return
	}

	go func() {
		for {
			time.Sleep(30 * time.Second)
			ws.WriteJSON(map[string]interface{}{"op": 1, "d": nil})
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		if opCode, ok := payload["op"].(float64); ok && opCode == 10 {
			if d, ok := payload["d"].(map[string]interface{}); ok {
				if interval, ok := d["heartbeat_interval"].(float64); ok {
					go func() {
						ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
						for range ticker.C {
							ws.WriteJSON(map[string]interface{}{"op": 1, "d": nil})
						}
					}()
				}
			}
		}
		if payload["t"] == "MESSAGE_REACTION_ADD" {
			if d, ok := payload["d"].(map[string]interface{}); ok {
				if userID, ok := d["user_id"].(string); ok && userID == TargetUserID {
					if emoji, ok := d["emoji"].(map[string]interface{}); ok {
						if emojiName, ok := emoji["name"].(string); ok && emojiName == TargetEmoji {
							channelID, _ := d["channel_id"].(string)
							messageID, _ := d["message_id"].(string)
							
							
							go func(cID, mID string) {
								reactUltraFast(cID, mID)
							}(channelID, messageID)
						}
					}
				}
			}
		}
	}
}

func reactUltraFast(channelID, messageID string) {
	emoji := url.QueryEscape(TargetEmoji)
	apiURL := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, emoji)

	req, _ := http.NewRequest("PUT", apiURL, nil)
	req.Header = http.Header{
		"Authorization":  {Token},
		"User-Agent":     {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"},
		"Content-Type":   {"application/json"},
		"Accept":         {"*/*"},
		"Cache-Control":  {"no-cache"},
		"Connection":     {"keep-alive"},
		"Pragma":         {"no-cache"},
	}

	client.Do(req)
}
