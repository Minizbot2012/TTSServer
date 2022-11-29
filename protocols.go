package ttsserver

import (
	"github.com/gorilla/websocket"
)

func SendTTSRequest(conn *websocket.Conn, Req string) error {
	request := &TTSRequest{
		Request: Req,
	}
	e := conn.WriteJSON(request)
	if e != nil {
		return e
	}
	return nil
}

func RecvTTSRequest(conn *websocket.Conn) (*TTSRequest, error) {
	req := &TTSRequest{}
	e := conn.ReadJSON(req)
	if e != nil {
		return nil, e
	}
	return req, nil
}

type TTSRequest struct {
	Request string
}

func SendTTSResponse(conn *websocket.Conn, Data []int16) error {
	response := &TTSResponse{
		TTSData: Data,
	}
	e := conn.WriteJSON(response)
	if e != nil {
		return e
	}
	return nil
}

func RecvTTSResponse(conn *websocket.Conn) (*TTSResponse, error) {
	resp := &TTSResponse{}
	e := conn.ReadJSON(resp)
	if e != nil {
		return nil, e
	}
	return resp, nil
}

type TTSResponse struct {
	TTSData []int16
}
