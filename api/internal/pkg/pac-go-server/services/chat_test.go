package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IBM/power-access-cloud/api/internal/pkg/pac-go-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func chatContext(userID, username string) context.Context {
	return getContext(testContext{userID: userID, username: username})
}

func makeSummaries(ended bool) []models.ConversationSummary {
	return []models.ConversationSummary{
		{
			ConversationID: 1,
			UserID:         "user1",
			Username:       "alice",
			MessageCount:   3,
			LastMessageAt:  time.Now(),
			Ended:          ended,
			FirstMessage:   "hello",
		},
	}
}

func makeMessages() []models.ChatMessage {
	return []models.ChatMessage{
		{
			ConversationID: 1,
			UserID:         "user1",
			Username:       "alice",
			Message:        "hello",
			Sender:         models.SenderUser,
			Timestamp:      time.Now(),
		},
		{
			ConversationID: 1,
			UserID:         "user1",
			Username:       "alice",
			Message:        "hi there",
			Sender:         models.SenderAdmin,
			Timestamp:      time.Now(),
		},
	}
}

// ginParam attaches URL params to a Gin context, matching how the router sets them.
func ginParam(c *gin.Context, key, value string) {
	c.Params = append(c.Params, gin.Param{Key: key, Value: value})
}

// ── GetUserConversations ──────────────────────────────────────────────────────

func TestGetUserConversations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		userID     string
		wantStatus int
	}{
		{
			name: "returns conversations for user",
			mockFunc: func() {
				mockDB.EXPECT().
					GetUserConversations(gomock.Any(), "user1").
					Return(makeSummaries(false), nil).Times(1)
			},
			userID:     "user1",
			wantStatus: http.StatusOK,
		},
		{
			name: "db error returns 500",
			mockFunc: func() {
				mockDB.EXPECT().
					GetUserConversations(gomock.Any(), "user1").
					Return(nil, errors.New("db down")).Times(1)
			},
			userID:     "user1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/pac-go-server/conversations", nil)
			c.Request = req.WithContext(chatContext(tc.userID, "alice"))
			dbCon = mockDB
			GetUserConversations(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// ── GetUserConversationMessages ───────────────────────────────────────────────

func TestGetUserConversationMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		convID     string
		wantStatus int
	}{
		{
			name: "returns messages for valid conv id",
			mockFunc: func() {
				mockDB.EXPECT().
					GetChatMessages(gomock.Any(), "user1", int64(1)).
					Return(makeMessages(), nil).Times(1)
			},
			convID:     "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid conv id returns 400",
			mockFunc:   func() {},
			convID:     "not-a-number",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "db error returns 500",
			mockFunc: func() {
				mockDB.EXPECT().
					GetChatMessages(gomock.Any(), "user1", int64(2)).
					Return(nil, errors.New("db error")).Times(1)
			},
			convID:     "2",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/pac-go-server/conversations/"+tc.convID+"/messages", nil)
			c.Request = req.WithContext(chatContext("user1", "alice"))
			ginParam(c, "conv_id", tc.convID)
			dbCon = mockDB
			GetUserConversationMessages(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// ── GetAdminConversations ─────────────────────────────────────────────────────

func TestGetAdminConversations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		wantStatus int
	}{
		{
			name: "returns all conversations",
			mockFunc: func() {
				mockDB.EXPECT().
					GetAllConversations(gomock.Any()).
					Return(makeSummaries(false), nil).Times(1)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "db error returns 500",
			mockFunc: func() {
				mockDB.EXPECT().
					GetAllConversations(gomock.Any()).
					Return(nil, errors.New("db down")).Times(1)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/pac-go-server/admin/conversations", nil)
			c.Request = req.WithContext(chatContext("admin1", "adminuser"))
			dbCon = mockDB
			GetAdminConversations(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// ── GetAdminConversationMessages ──────────────────────────────────────────────

func TestGetAdminConversationMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		userID     string
		convID     string
		wantStatus int
	}{
		{
			name: "returns messages for valid user and conv",
			mockFunc: func() {
				mockDB.EXPECT().
					GetChatMessages(gomock.Any(), "user1", int64(1)).
					Return(makeMessages(), nil).Times(1)
			},
			userID:     "user1",
			convID:     "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid conv id returns 400",
			mockFunc:   func() {},
			userID:     "user1",
			convID:     "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "db error returns 500",
			mockFunc: func() {
				mockDB.EXPECT().
					GetChatMessages(gomock.Any(), "user1", int64(1)).
					Return(nil, errors.New("db error")).Times(1)
			},
			userID:     "user1",
			convID:     "1",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			c.Request = req.WithContext(chatContext("admin1", "adminuser"))
			ginParam(c, "user_id", tc.userID)
			ginParam(c, "conv_id", tc.convID)
			dbCon = mockDB
			GetAdminConversationMessages(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// ── AdminReply ────────────────────────────────────────────────────────────────

func TestAdminReply(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		adminID    string
		userID     string
		convID     string
		body       interface{}
		wantStatus int
	}{
		{
			name: "reply enqueued successfully",
			mockFunc: func() {
				// InsertChatMessage is called asynchronously by the worker;
				// the handler may or may not call it directly during the test
				// window, so we allow 0-1 calls.
				mockDB.EXPECT().
					InsertChatMessage(gomock.Any(), gomock.Any()).
					Return(nil).AnyTimes()
			},
			adminID:    "admin1",
			userID:     "user1",
			convID:     "1",
			body:       map[string]string{"message": "hello from admin"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing message body returns 400",
			mockFunc:   func() {},
			adminID:    "admin1",
			userID:     "user1",
			convID:     "1",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid conv id returns 400",
			mockFunc:   func() {},
			adminID:    "admin1",
			userID:     "user1",
			convID:     "nan",
			body:       map[string]string{"message": "hi"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing admin id returns 401",
			mockFunc:   func() {},
			adminID:    "", // no auth
			userID:     "user1",
			convID:     "1",
			body:       map[string]string{"message": "hi"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Drain the reply queue before each test so queue-full state doesn't leak.
			for len(adminReplyQueue) > 0 {
				<-adminReplyQueue
			}

			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, err := json.Marshal(tc.body)
			require.NoError(t, err)
			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req.WithContext(chatContext(tc.adminID, "adminuser"))
			ginParam(c, "user_id", tc.userID)
			ginParam(c, "conv_id", tc.convID)
			dbCon = mockDB
			AdminReply(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// TestAdminReply_QueueFull tests the synchronous-write fallback when the
// adminReplyQueue is at capacity. It swaps in a zero-capacity channel for the
// duration of the test and restores it afterwards.
func TestAdminReply_QueueFull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	// Swap the package-level queue for a zero-buffer one so the select default
	// branch fires deterministically without racing the background worker.
	orig := adminReplyQueue
	adminReplyQueue = make(chan *models.ChatMessage)
	defer func() { adminReplyQueue = orig }()

	mockDB.EXPECT().
		InsertChatMessage(gomock.Any(), gomock.Any()).
		Return(errors.New("db write failed")).Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]string{"message": "overflow"})
	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req.WithContext(chatContext("admin1", "adminuser"))
	ginParam(c, "user_id", "user1")
	ginParam(c, "conv_id", "1")
	dbCon = mockDB
	AdminReply(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ── AdminEndConversation ──────────────────────────────────────────────────────

func TestAdminEndConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	tests := []struct {
		name       string
		mockFunc   func()
		userID     string
		convID     string
		wantStatus int
	}{
		{
			name: "ends conversation successfully",
			mockFunc: func() {
				mockDB.EXPECT().
					MarkConversationEnded(gomock.Any(), "user1", "", int64(1)).
					Return(nil).Times(1)
			},
			userID:     "user1",
			convID:     "1",
			wantStatus: http.StatusOK,
		},
		{
			// DB failure is non-fatal — the hub is still notified and 200 returned.
			name: "db failure is non-fatal — still returns 200",
			mockFunc: func() {
				mockDB.EXPECT().
					MarkConversationEnded(gomock.Any(), "user1", "", int64(1)).
					Return(errors.New("db error")).Times(1)
			},
			userID:     "user1",
			convID:     "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid conv id returns 400",
			mockFunc:   func() {},
			userID:     "user1",
			convID:     "bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodPost, "/", nil)
			c.Request = req.WithContext(chatContext("admin1", "adminuser"))
			ginParam(c, "user_id", tc.userID)
			ginParam(c, "conv_id", tc.convID)
			dbCon = mockDB
			AdminEndConversation(c)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// ── chatHub unit tests ────────────────────────────────────────────────────────

func TestChatHub_PublishToSubscribers(t *testing.T) {
	h := &chatHub{subs: make(map[string][]chan hubMessage)}
	ch1 := make(chan hubMessage, 1)
	ch2 := make(chan hubMessage, 1)

	unsub1 := h.subscribe("key1", ch1)
	unsub2 := h.subscribe("key1", ch2)

	msg := hubMessage{ConversationID: 1, Sender: models.SenderUser, Message: "hi"}
	h.publish("key1", msg)

	assert.Equal(t, msg, <-ch1, "ch1 should receive the published message")
	assert.Equal(t, msg, <-ch2, "ch2 should receive the published message")

	unsub1()
	unsub2()
}

func TestChatHub_NoDeliveryAfterUnsubscribe(t *testing.T) {
	h := &chatHub{subs: make(map[string][]chan hubMessage)}
	ch := make(chan hubMessage, 1)

	unsub := h.subscribe("key1", ch)
	unsub() // unsubscribe immediately

	h.publish("key1", hubMessage{Message: "ghost"})

	select {
	case m := <-ch:
		t.Fatalf("expected no message after unsub, got %v", m)
	default:
		// correct — nothing delivered
	}
}

func TestChatHub_KeyDeletedWhenEmpty(t *testing.T) {
	h := &chatHub{subs: make(map[string][]chan hubMessage)}
	ch := make(chan hubMessage, 1)

	unsub := h.subscribe("key1", ch)
	h.mu.Lock()
	_, exists := h.subs["key1"]
	h.mu.Unlock()
	assert.True(t, exists, "key should exist after subscribe")

	unsub()
	h.mu.Lock()
	_, exists = h.subs["key1"]
	h.mu.Unlock()
	assert.False(t, exists, "key should be deleted after last unsub")
}

func TestChatHub_DropsMessageWhenBufferFull(t *testing.T) {
	h := &chatHub{subs: make(map[string][]chan hubMessage)}
	// Buffer size 1 — fill it so the next publish triggers the drop path.
	ch := make(chan hubMessage, 1)
	ch <- hubMessage{Message: "blocker"}
	unsub := h.subscribe("key1", ch)
	defer unsub()

	h.publish("key1", hubMessage{Message: "dropped"})

	assert.Equal(t, int64(1), h.dropCount.Load(), "drop counter should be 1")
	// Original message is still in the buffer, not the dropped one.
	assert.Equal(t, "blocker", (<-ch).Message)
}

func TestChatHub_MultipleKeys_Isolated(t *testing.T) {
	h := &chatHub{subs: make(map[string][]chan hubMessage)}
	chA := make(chan hubMessage, 1)
	chB := make(chan hubMessage, 1)

	unsubA := h.subscribe("keyA", chA)
	unsubB := h.subscribe("keyB", chB)
	defer unsubA()
	defer unsubB()

	h.publish("keyA", hubMessage{Message: "for A"})

	select {
	case m := <-chA:
		assert.Equal(t, "for A", m.Message)
	default:
		t.Fatal("chA should have received message")
	}
	select {
	case m := <-chB:
		t.Fatalf("chB should not have received message from keyA, got %v", m)
	default:
		// correct
	}
}

func TestUserToAdminKey(t *testing.T) {
	assert.Equal(t, "abc:42", userToAdminKey("abc", 42))
	assert.Equal(t, "xyz:1", userToAdminKey("xyz", 1))
}

// ── StartAdminReplyWorker ─────────────────────────────────────────────────────

func TestStartAdminReplyWorker_PersistsMessages(t *testing.T) {
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	dbCon = mockDB
	mockDB.EXPECT().
		InsertChatMessage(gomock.Any(), gomock.Any()).
		Return(nil).Times(2)

	// Drain any leftover items from earlier tests.
	for len(adminReplyQueue) > 0 {
		<-adminReplyQueue
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Enqueue two messages before starting the worker.
	adminReplyQueue <- &models.ChatMessage{UserID: "u1", ConversationID: 1, Message: "m1"}
	adminReplyQueue <- &models.ChatMessage{UserID: "u1", ConversationID: 1, Message: "m2"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		StartAdminReplyWorker(ctx)
	}()

	// Give the worker time to drain the two messages, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done
}

func TestStartAdminReplyWorker_DrainOnShutdown(t *testing.T) {
	_, mockDB, _, tearDown := setUp(t)
	defer tearDown()

	dbCon = mockDB
	// One message already in the queue when shutdown fires — must still be persisted.
	mockDB.EXPECT().
		InsertChatMessage(gomock.Any(), gomock.Any()).
		Return(nil).Times(1)

	for len(adminReplyQueue) > 0 {
		<-adminReplyQueue
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so the worker enters shutdown drain

	adminReplyQueue <- &models.ChatMessage{UserID: "u1", ConversationID: 1, Message: "drain me"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		StartAdminReplyWorker(ctx)
	}()

	select {
	case <-done:
		// worker exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit within timeout")
	}
}
