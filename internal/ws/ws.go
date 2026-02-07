package ws

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"strings"
	"wstogo/internal/http"
)

type WSConn struct {
	conn net.Conn
}

func NewWSConn(addr string) (*WSConn, error) {
	return &WSConn{}, nil
}

func (w *WSConn) Handshake(host, path string) (error, string) {
	u, _ := url.Parse("ws://" + host + path)
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err, ""
	}
	w.conn = conn

	key := make([]byte, 16)
	rand.Read(key)
	secKey := base64.StdEncoding.EncodeToString(key)
	req := http.NewRequest("GET", path, "HTTP/1.1", nil)
	req.AddHeader("Host", u.Host)
	req.AddHeader("Upgrade", "websocket")
	req.AddHeader("Connection", "Upgrade")
	req.AddHeader("Sec-WebSocket-Key", secKey)
	req.AddHeader("Sec-WebSocket-Version", "13")
	req.AddHeader("Origin", "http://localhost")

	fmt.Fprintf(w.conn, "%s", req.String())

	// capture full response headers for TUI log
	respReader := bufio.NewReader(w.conn)
	respLines := []string{}
	line, _ := respReader.ReadString('\n')
	statusLine := strings.TrimSpace(line)
	respLines = append(respLines, line)
	parts := strings.Split(statusLine, " ")
	code := 0
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &code)
	}
	if code != 101 {
		return fmt.Errorf("upgrade failed: %s", statusLine), ""
	}

	tp := textproto.NewReader(respReader)
	mimeHeader, _ := tp.ReadMIMEHeader()
	for k, vs := range mimeHeader {
		if len(vs) > 0 {
			respLines = append(respLines, k+": "+vs[0]+"\r\n")
		}
	}
	respLines = append(respLines, "\r\n")
	resStr := strings.Join(respLines, "")

	return nil, resStr
}

func (w *WSConn) ReadMessage() (string, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(w.conn, b)
	if err != nil {
		return "", err
	}
	if b[0]&0x0F != 1 {
		return "", fmt.Errorf("non-text")
	}
	length := int(b[1] & 0x7F)
	data := make([]byte, length)
	_, err = io.ReadFull(w.conn, data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (w *WSConn) WriteMessage(msg string) error {
	payload := []byte(msg)
	// client must mask: 4-byte key, set bit, XOR
	maskKey := make([]byte, 4)
	rand.Read(maskKey)
	masked := make([]byte, len(payload))
	for i := range payload {
		masked[i] = payload[i] ^ maskKey[i%4]
	}
	// header: FIN+text, mask bit + len
	header := []byte{0x81, 0x80 | byte(len(payload))}
	header = append(header, maskKey...)
	header = append(header, masked...)
	_, err := w.conn.Write(header)
	return err
}

func (w *WSConn) Close() {
	if w.conn != nil {
		w.conn.Close()
	}
}
