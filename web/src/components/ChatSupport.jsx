import React, { useState, useEffect, useRef } from "react";
import { Search, Send } from "@carbon/icons-react";
import UserService from "../services/UserService";
import "../styles/chat-support.scss";

const ChatSupport = () => {
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState("");
  const [conversationId, setConversationId] = useState(null);
  const [ws, setWs] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState("connecting");
  const messagesEndRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const getWebSocketUrl = () => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const token = UserService.getToken();
    return `${protocol}//${host}/pac-go-server/chat?token=${encodeURIComponent(token)}`;
  };

  const connectWebSocket = () => {
    try {
      const wsUrl = getWebSocketUrl();
      console.log("Connecting to WebSocket:", wsUrl.replace(/token=[^&]+/, "token=***"));
      
      const websocket = new WebSocket(wsUrl);

      websocket.onopen = () => {
        console.log("WebSocket connected");
        setConnectionStatus("connected");
      };

      websocket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log("Received message:", data);

          // Set conversation ID from first response
          if (data.conversation_id && !conversationId) {
            setConversationId(`conv-${data.conversation_id}`);
          }

          // Add admin message to chat
          if (data.message && data.sender === "admin") {
            const adminMessage = {
              id: Date.now(),
              sender: "Admin",
              content: data.message,
              timestamp: data.timestamp ? new Date(data.timestamp) : new Date(),
            };
            setMessages((prev) => [...prev, adminMessage]);
          }
        } catch (error) {
          console.error("Error parsing message:", error);
        }
      };

      websocket.onerror = (error) => {
        console.error("WebSocket error:", error);
        setConnectionStatus("error");
      };

      websocket.onclose = () => {
        console.log("WebSocket disconnected");
        setConnectionStatus("disconnected");
        
        // Auto-reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log("Attempting to reconnect...");
          connectWebSocket();
        }, 3000);
      };

      setWs(websocket);
    } catch (error) {
      console.error("Error creating WebSocket:", error);
      setConnectionStatus("error");
    }
  };

  useEffect(() => {
    connectWebSocket();

    // Cleanup on unmount
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (ws) {
        ws.close();
      }
    };
  }, []);

  const formatTime = (date) => {
    return date.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: true,
    });
  };

  const handleSendMessage = () => {
    if (inputMessage.trim() === "" || !ws || ws.readyState !== WebSocket.OPEN) {
      console.warn("Cannot send message: WebSocket not ready");
      return;
    }

    const message = {
      message: inputMessage,
    };

    try {
      ws.send(JSON.stringify(message));
      console.log("Message sent:", message);

      // Add user message to UI immediately
      const userMessage = {
        id: Date.now(),
        sender: "You",
        content: inputMessage,
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, userMessage]);
      setInputMessage("");
    } catch (error) {
      console.error("Error sending message:", error);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const getConnectionStatusText = () => {
    switch (connectionStatus) {
      case "connected":
        return "Connected";
      case "connecting":
        return "Connecting...";
      case "disconnected":
        return "Disconnected (reconnecting...)";
      case "error":
        return "Connection Error";
      default:
        return "Unknown";
    }
  };

  return (
    <div className="chat-support-container">
      {/* Left Panel - Conversations */}
      <div className="conversations-panel">
        <div className="conversations-header">
          <h3>Conversations</h3>
          <div className="search-box">
            <Search size={16} />
            <input type="text" placeholder="Search" />
          </div>
        </div>
        <div className="conversations-list">
          <div className="conversation-item active">
            <div className="conversation-avatar">
              <div className="avatar-circle">U</div>
            </div>
            <div className="conversation-details">
              <div className="conversation-id">
                {conversationId || "Your first conversation"}
              </div>
              <div className="conversation-status">
                {getConnectionStatusText()}
              </div>
              <div className="conversation-unread">
                {messages.length > 0 ? `${messages.length} messages` : "Start chatting"}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Right Panel - Chat Area */}
      <div className="chat-panel">
        <div className="chat-header">
          <div className="chat-header-content">
            <h2>Conversation ID: {conversationId || "Initializing..."}</h2>
            <div className="chat-status">
              <span className="status-label">
                Conversation ID: {conversationId || "Initializing..."}
              </span>
              <span className="status-divider">|</span>
              <span className={`status-${connectionStatus}`}>
                Status: {getConnectionStatusText()}
              </span>
            </div>
          </div>
        </div>

        <div className="chat-messages">
          {messages.length === 0 ? (
            <div className="empty-state">
              <p>Start your conversation by typing a message below</p>
            </div>
          ) : (
            messages.map((message) => (
              <div
                key={message.id}
                className={`message ${
                  message.sender === "You" ? "message-user" : "message-admin"
                }`}
              >
                <div className="message-content">
                  <div className="message-header">
                    <span className="message-sender">
                      [{message.sender}, {formatTime(message.timestamp)}]:
                    </span>
                  </div>
                  <div className="message-text">{message.content}</div>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="chat-input-container">
          <div className="chat-input-wrapper">
            <input
              type="text"
              className="chat-input"
              placeholder="Type your message..."
              value={inputMessage}
              onChange={(e) => setInputMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              disabled={connectionStatus !== "connected"}
            />
            <button
              className="send-button"
              onClick={handleSendMessage}
              disabled={
                inputMessage.trim() === "" || connectionStatus !== "connected"
              }
            >
              <Send size={20} />
              <span>Send</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatSupport;
