package ws

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Arif14377/golang-simple-messaging-app/app/middleware"
	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/dto"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"go.elastic.co/apm/v2"
)

const (
	// WSMessagePath adalah route WebSocket chat.
	WSMessagePath = "/message/v1/send"
	// DefaultRoomID adalah room global tempat semua pengguna terhubung.
	DefaultRoomID = "lobby"
	// MaxMessageLength harus sama dengan maxlength textarea di frontend.
	MaxMessageLength = 2000

	readLimit       = 16 << 10 // 16 KiB, cukup untuk pesan 2000 karakter
	writeWait       = 10 * time.Second
	saveTimeout     = 5 * time.Second
	broadcastBuffer = 256
)

// incomingMessage adalah payload yang dikirim client. Identitas pengirim
// sengaja tidak dibaca dari client: user id & username diambil dari JWT cookie.
type incomingMessage struct {
	RoomID  string `json:"roomId"`
	Message string `json:"message"`
}

// errorMessage dipakai saat pesan dari client ditolak. Hanya dikirim ke
// pengirimnya, bukan di-broadcast.
type errorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// client adalah satu koneksi WebSocket beserta identitas pemiliknya.
type client struct {
	conn     *websocket.Conn
	userID   uint
	username string
}

// Packet internal hub: target nil berarti broadcast ke semua client.
type packet struct {
	target  *client
	payload any
}

type hub struct {
	mu        sync.RWMutex
	clients   map[*client]struct{}
	broadcast chan packet
}

func newHub() *hub {
	return &hub{
		clients:   make(map[*client]struct{}),
		broadcast: make(chan packet, broadcastBuffer),
	}
}

// RegisterWSRoutes mendaftarkan route chat pada app yang sama dengan HTTP/API.
// Semua koneksi berbagi satu hub; route dilindungi middleware.Auth sehingga
// hanya pengguna dengan sesi valid yang bisa terhubung.
func RegisterWSRoutes(app *fiber.App) {
	h := newHub()
	go h.run()

	app.Get(WSMessagePath, middleware.Auth, websocket.New(h.handle))
}

// run adalah satu-satunya goroutine yang menulis ke socket, sehingga tidak ada
// penulisan bersamaan ke koneksi yang sama.
func (h *hub) run() {
	for pkt := range h.broadcast {
		h.dispatch(pkt)
	}
}

// dispatch mengirim satu packet dan mencatatnya sebagai transaksi APM sendiri.
//
// Ini perlu karena processMessage sudah menutup transaksi "ws.message" sebelum
// packet sampai ke goroutine ini. Span yang dibuat tanpa transaksi aktif akan
// dibuang oleh agent, jadi broadcast harus punya transaksi sendiri.
func (h *hub) dispatch(pkt packet) {
	ctx, tx := tracing.StartTransaction(context.Background(), "ws.broadcast", "websocket")
	if tx == nil {
		h.deliver(context.Background(), pkt)
		return
	}

	result, outcome := "sent", "success"
	defer func() {
		tx.Result = result
		tx.Outcome = outcome
		tx.End()
	}()

	written, failed := h.deliver(ctx, pkt)
	tx.Context.SetLabel("targets", strconv.Itoa(written+failed))
	if failed > 0 {
		tx.Context.SetLabel("failed_writes", strconv.Itoa(failed))
		outcome = "failure"
		result = "failed"
		if written > 0 {
			result = "partial"
		}
	}
}

// deliver menulis payload ke target tertentu, atau ke semua client jika
// pkt.target nil. Nilai balik dipakai dispatch untuk menentukan hasil transaksi.
func (h *hub) deliver(ctx context.Context, pkt packet) (written, failed int) {
	if pkt.target != nil {
		if h.write(ctx, pkt.target, pkt.payload) {
			return 1, 0
		}
		return 0, 1
	}

	for _, c := range h.snapshot() {
		if h.write(ctx, c, pkt.payload) {
			written++
		} else {
			failed++
		}
	}

	return written, failed
}

func (h *hub) handle(conn *websocket.Conn) {
	c := &client{
		conn:     conn,
		userID:   localUint(conn, "userId"),
		username: localString(conn, "username"),
	}
	if c.userID == 0 || c.username == "" {
		_ = conn.WriteJSON(errorMessage{
			Type:    "error",
			Message: "Sesi tidak valid, silakan login ulang.",
		})
		_ = conn.Close()
		return
	}

	h.add(c)
	defer func() {
		h.remove(c)
		_ = conn.Close()
	}()

	conn.SetReadLimit(readLimit)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			// client menutup koneksi / koneksi terputus.
			return
		}

		h.processMessage(c, data)
	}
}

// processMessage menangani satu pesan: validasi, simpan ke Mongo, lalu
// broadcast. Setiap pesan dibungkus transaksi APM sendiri agar latensinya
// terukur per pesan, bukan per koneksi yang bisa hidup berjam-jam.
func (h *hub) processMessage(c *client, data []byte) {
	ctx, tx := tracing.StartTransaction(context.Background(), "ws.message", "websocket")
	result, outcome := "rejected", "unknown"
	if tx != nil {
		defer func() {
			tx.Result = result
			tx.Outcome = outcome
			tx.End()
		}()
		tx.Context.SetUserID(strconv.FormatUint(uint64(c.userID), 10))
		tx.Context.SetUsername(c.username)
	}

	var incoming incomingMessage
	if err := json.Unmarshal(data, &incoming); err != nil {
		h.sendError(c, "Format pesan tidak valid.")
		return
	}

	content := strings.TrimSpace(incoming.Message)
	if content == "" {
		h.sendError(c, "Pesan tidak boleh kosong.")
		return
	}
	if utf8.RuneCountInString(content) > MaxMessageLength {
		h.sendError(c, "Pesan terlalu panjang (maksimal 2000 karakter).")
		return
	}

	message := models.Message{
		RoomID:    normalizeRoomID(incoming.RoomID),
		UserID:    c.userID,
		Username:  c.username,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	if err := message.Validate(ctx); err != nil {
		log.Printf("Invalid chat message: %v\n", err)
		h.sendError(c, "Pesan tidak valid.")
		return
	}

	if tx != nil {
		tx.Context.SetLabel("room_id", message.RoomID)
	}

	saveCtx, cancel := context.WithTimeout(ctx, saveTimeout)
	err := repository.InsertMessage(saveCtx, &message)
	cancel()
	if err != nil {
		log.Printf("Failed to save chat message: %v\n", err)
		// jangan broadcast: riwayat Mongo adalah sumber kebenaran.
		h.sendError(c, "Pesan gagal disimpan, coba lagi.")
		result, outcome = "save_failed", "failure"
		return
	}

	h.broadcast <- packet{payload: dto.NewMessageResponse(message)}
	result, outcome = "saved", "success"
}

func (h *hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *hub) snapshot() []*client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}

	return clients
}

func (h *hub) sendError(c *client, message string) {
	h.broadcast <- packet{
		target: c,
		payload: errorMessage{
			Type:    "error",
			Message: message,
		},
	}
}

// write mengirim satu payload. Dipanggil hanya dari goroutine run. Nilai
// kembali false berarti koneksi sudah tidak bisa dipakai dan sudah ditutup.
func (h *hub) write(ctx context.Context, c *client, payload any) bool {
	ctx, span := tracing.StartSpan(ctx, "ws.write", "websocket")
	defer span.End()
	span.Context.SetLabel("user_id", strconv.FormatUint(uint64(c.userID), 10))

	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteJSON(payload); err != nil {
		span.Outcome = "failure"
		apm.CaptureError(ctx, err).Send()
		h.remove(c)
		_ = c.conn.Close()
		return false
	}

	return true
}

func normalizeRoomID(roomID string) string {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return DefaultRoomID
	}

	return roomID
}

func localUint(conn *websocket.Conn, key string) uint {
	value, ok := conn.Locals(key).(uint)
	if !ok {
		return 0
	}

	return value
}

func localString(conn *websocket.Conn, key string) string {
	value, ok := conn.Locals(key).(string)
	if !ok {
		return ""
	}

	return value
}
