package podapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"learn-go/k8s-api/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 前端发来的消息
// 输入: {"type":"input",  "data":"ls\n"}
// 调整: {"type":"resize", "rows":24, "cols":80}
type TermMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// TerminalSizeQueue 通知 Pod 终端尺寸变化
type TerminalSizeQueue struct {
	ch chan remotecommand.TerminalSize
}

func (t *TerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-t.ch
	if !ok {
		return nil
	}
	return &size
}

// chanReader 把 channel 包装成 io.Reader，供 client-go 读取用户输入
type chanReader struct {
	ch  chan []byte
	buf []byte
}

func (r *chanReader) Read(p []byte) (int, error) {
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}
	data, ok := <-r.ch
	if !ok {
		return 0, io.EOF
	}
	n := copy(p, data)
	r.buf = data[n:]
	return n, nil
}

// wsWriter 把 Pod 输出写回 WebSocket（加锁防止并发写）
type wsWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// ExecPod GET /api/pods/:namespace/:pod/exec?container=nginx
func ExecPod(c *gin.Context) {
	namespace := c.Param("namespace")
	pod       := c.Param("pod")
	container := c.Query("container")

	// 1. HTTP 升级为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 2. 创建通信 channel
	inputCh := make(chan []byte)
	sizeCh  := make(chan remotecommand.TerminalSize, 1)

	// 3. goroutine 持续读取前端消息
	go func() {
		defer close(inputCh)
		defer close(sizeCh)
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg TermMsg
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "input":
				inputCh <- []byte(msg.Data)
			case "resize":
				select {
				case sizeCh <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}:
				default:
				}
			}
		}
	}()

	// 4. 构造 kubectl exec 请求
	req := config.Client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/sh"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 5. 创建 SPDY 执行器
	executor, err := remotecommand.NewSPDYExecutor(config.RestConfig, "POST", req.URL())
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("exec 失败: "+err.Error()))
		return
	}

	// 6. 桥接 WebSocket ↔ Pod（阻塞直到终端关闭）
	writer := &wsWriter{conn: conn}
	err = executor.StreamWithContext(c.Request.Context(), remotecommand.StreamOptions{
		Stdin:             &chanReader{ch: inputCh},
		Stdout:            writer,
		Stderr:            writer,
		Tty:               true,
		TerminalSizeQueue: &TerminalSizeQueue{ch: sizeCh},
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("\r\n[连接断开]\r\n"))
	}
}
