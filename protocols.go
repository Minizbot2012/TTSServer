package ttsserver

import (
	"io"

	"github.com/Minizbot2012/minxdr"
)

func SendTTSRequest(conn io.Writer, Req string) error {
	request := &TTSRequest{
		Request: Req,
	}
	_, err := minxdr.Marshal(conn, request)
	if err != nil {
		return err
	}
	return nil
}

func RecvTTSRequest(conn io.Reader) (*TTSRequest, error) {
	req := new(TTSRequest)
	_, e := minxdr.Unmarshal(conn, req)
	if e != nil {
		println("REQ: DEC: " + e.Error())
		return nil, e
	}
	return req, nil
}

type TTSRequest struct {
	Request string
}

func SendTTSResponse(conn io.Writer, Data []int16) error {
	response := &TTSResponse{
		TTSData: Data,
	}
	_, err := minxdr.Marshal(conn, response)
	if err != nil {
		return err
	}
	return nil
}

func RecvTTSResponse(conn io.Reader) (*TTSResponse, error) {
	resp := new(TTSResponse)
	_, e := minxdr.Unmarshal(conn, resp)
	if e != nil {
		return nil, e
	}
	return resp, nil
}

type TTSResponse struct {
	TTSData []int16
}
