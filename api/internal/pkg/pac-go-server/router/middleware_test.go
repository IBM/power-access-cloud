package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTokenFromWebSocketSubprotocols(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"valid", "pac.v1, abc.def.ghi", "abc.def.ghi"},
		{"valid without space", "pac.v1,abc.def.ghi", "abc.def.ghi"},
		{"extra whitespace", "  pac.v1 ,\t abc.def.ghi  ", "abc.def.ghi"},
		{"marker case-insensitive", "PAC.V1, abc.def.ghi", "abc.def.ghi"},
		{"empty token", "pac.v1, ", ""},
		{"trailing comma", "pac.v1,", ""},
		{"empty", "", ""},
		{"marker only", "pac.v1", ""},
		{"wrong marker", "other, abc.def.ghi", ""},
		{"token first", "abc.def.ghi, pac.v1", ""},
		{"too many items", "pac.v1, abc, def", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tokenFromWebSocketSubprotocols(tt.header))
		})
	}
}

func TestInjectTokenFromSubprotocol(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testcases := []struct {
		name         string
		target       string
		headers      map[string]string
		wantAuth     string
		wantSubproto string
	}{
		{
			name:         "promotes subprotocol token and scrubs it from the header",
			target:       "/ws",
			headers:      map[string]string{"Sec-WebSocket-Protocol": "pac.v1, tok"},
			wantAuth:     "Bearer tok",
			wantSubproto: "pac.v1",
		},
		{
			name:         "ignores query string token",
			target:       "/ws?token=tok",
			headers:      nil,
			wantAuth:     "",
			wantSubproto: "",
		},
		{
			name:   "does not override existing Authorization",
			target: "/ws",
			headers: map[string]string{
				"Authorization":          "Bearer existing",
				"Sec-WebSocket-Protocol": "pac.v1, tok",
			},
			wantAuth:     "Bearer existing",
			wantSubproto: "pac.v1",
		},
		{
			name:         "leaves unrelated subprotocol headers untouched",
			target:       "/ws",
			headers:      map[string]string{"Sec-WebSocket-Protocol": "chat, superchat"},
			wantAuth:     "",
			wantSubproto: "chat, superchat",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var auth, subprotocol string
			r := gin.New()
			r.GET("/ws", InjectTokenFromSubprotocol, func(c *gin.Context) {
				auth = c.GetHeader("Authorization")
				subprotocol = c.GetHeader("Sec-WebSocket-Protocol")
			})
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(httptest.NewRecorder(), req)
			assert.Equal(t, tc.wantAuth, auth)
			assert.Equal(t, tc.wantSubproto, subprotocol)
		})
	}
}
