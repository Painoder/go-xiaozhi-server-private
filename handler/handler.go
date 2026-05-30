package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/xdimtech/go-xiaozhi/handler/base"
	"github.com/xdimtech/go-xiaozhi/handler/glmrealtime"
	"github.com/xdimtech/go-xiaozhi/handler/openai"
	"github.com/xdimtech/go-xiaozhi/handler/xiaozhi"
	"github.com/xdimtech/go-xiaozhi/pkg/config"

	"github.com/gorilla/websocket"
	"github.com/xdimtech/go-xiaozhi/pkg/utils"
)

type WebSocketServer struct {
	requestCounter atomic.Int64
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{}
}

func (s *WebSocketServer) Start(addr string) error {
	http.HandleFunc("/xiaozhi/v1/", s.RealTime)
	http.HandleFunc("/xiaozhi/ota/", s.HandleOTA)
	log.Printf("Server started at local: ws://127.0.0.1%s\n", addr)
	ip, _ := utils.GetLocalIP()
	log.Printf("Server started at public: ws://%s%s\n", ip, addr)
	return http.ListenAndServe(addr, nil)
}

func (s *WebSocketServer) wsConnect(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *WebSocketServer) RealTime(w http.ResponseWriter, r *http.Request) {

	var err error
	conn, err := s.wsConnect(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		http.Error(w, "websocket upgrade required", http.StatusBadRequest)
		return
	}

	defer func() {
		_ = conn.Close()
		conn = nil
	}()

	ctx := r.Context()
	connWrapper, err := s.NewConnWrapper(ctx, conn, r)
	if err != nil {
		log.Printf("create connection wrapper failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_ = connWrapper.ReadLoop(ctx)
}

func (s *WebSocketServer) NewConnWrapper(
	ctx context.Context, conn *websocket.Conn, r *http.Request) (base.WsConnWrapper, error) {
	if config.Provider().Name == "openai" {
		return openai.NewConnWrapper(ctx, conn)
	}
	if config.Provider().Name == "glm-realtime" {
		return glmrealtime.NewConnWrapper(ctx, conn)
	}
	return xiaozhi.NewConnWrapper(ctx, conn, xiaozhi.WithOriginReq(r))
}

type OTAResponse struct {
	ServerTime OTAServerTime `json:"server_time"`
	Firmware   OTAFirmware   `json:"firmware"`
	Websocket  *OTAWebsocket `json:"websocket,omitempty"`
}

type OTAServerTime struct {
	Timestamp      int64 `json:"timestamp"`
	TimezoneOffset int   `json:"timezone_offset"`
}

type OTAFirmware struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}

type OTAWebsocket struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func (s *WebSocketServer) HandleOTA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		ip, _ := utils.GetLocalIP()
		wsURL := "ws://" + ip + ":8000/xiaozhi/v1/"
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OTA接口运行正常，向设备发送的websocket地址是：" + wsURL))
		return
	}

	if r.Method == http.MethodPost {
		deviceID := r.Header.Get("device-id")
		clientID := r.Header.Get("client-id")
		if deviceID == "" {
			log.Printf("OTA POST: missing device-id header")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "missing device-id"})
			return
		}
		if clientID == "" {
			log.Printf("OTA POST: missing client-id header")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "missing client-id"})
			return
		}

		log.Printf("OTA POST: device-id=%s, client-id=%s", deviceID, clientID)

		ip, _ := utils.GetLocalIP()
		wsURL := "ws://" + ip + ":8000/xiaozhi/v1/"

		deviceVersion := r.Header.Get("device-version")
		if deviceVersion == "" {
			deviceVersion = r.Header.Get("firmware-version")
		}
		if deviceVersion == "" {
			deviceVersion = r.Header.Get("app-version")
		}
		if deviceVersion == "" {
			deviceVersion = "0.0.0"
		}

		resp := OTAResponse{
			ServerTime: OTAServerTime{
				Timestamp:      time.Now().UnixMilli(),
				TimezoneOffset: 480,
			},
			Firmware: OTAFirmware{
				Version: deviceVersion,
				URL:     "",
			},
			Websocket: &OTAWebsocket{
				URL:   wsURL,
				Token: "",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
