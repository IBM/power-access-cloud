import React, { useState, useEffect, useRef, useCallback } from "react";
import { Send, Renew } from "@carbon/icons-react";
import {
  Button,
  DataTableSkeleton,
  InlineNotification,
  TextArea,
} from "@carbon/react";
import "../styles/chat-support.scss";
import axios from "axios";
import UserService from "../services/UserService";

const api = axios.create();
api.interceptors.request.use((config) => {
  const cb = () => {
    config.headers.Authorization = `Bearer ${UserService.getToken()}`;
    return Promise.resolve(config);
  };
  return UserService.updateToken(cb);
});

// Returns "username-conv-N" if username is available, else falls back to userID.
const convLabel = (username, userID, convID) =>
  `${username || userID}-conv-${convID}`;

// Unique key for a (userID, convID) pair — used for unread tracking.
const convKey = (userID, convID) => `${userID}__${convID}`;

const ChatAdmin = () => {
  const [conversations, setConversations] = useState([]);
  const [selected, setSelected] = useState(null); // { userID, convID }
  const [messages, setMessages] = useState([]);
  const [replyText, setReplyText] = useState("");
  const [loading, setLoading] = useState(false);
  const [msgLoading, setMsgLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [convEnded, setConvEnded] = useState(false); // true after conversation ended
  // unreadCounts: { [convKey]: number } — count of unread user messages per conv.
  const [unreadCounts, setUnreadCounts] = useState({});

  // Ref to the live-watch WebSocket for the currently selected conversation.
  const watchWsRef = useRef(null);
  const messagesEndRef = useRef(null);
  // Ref copy of selected so WS onmessage can read it without a stale closure.
  const selectedRef = useRef(null);

  // Auto-scroll to the latest message whenever the list changes.
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Keep selectedRef in sync with selected state.
  useEffect(() => {
    selectedRef.current = selected;
  }, [selected]);

  const loadConversations = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await api.get("/pac-go-server/admin/conversations");
      setConversations(res.data.conversations ?? []);
    } catch (err) {
      setError("Failed to load conversations: " + (err.response?.data?.error ?? err.message));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadConversations();
  }, [loadConversations]);

  // Close any existing watch WebSocket cleanly.
  const closeWatchWs = () => {
    const ws = watchWsRef.current;
    if (ws) {
      ws.onclose = null;
      ws.onerror = null;
      ws.onmessage = null;
      ws.close(1000, "conversation changed");
      watchWsRef.current = null;
    }
  };

  // Open a watch WebSocket for (userID, convID) and append incoming messages.
  const openWatchWs = (userID, convID) => {
    closeWatchWs();

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const token = UserService.getToken();
    const url = `${protocol}//${window.location.host}/pac-go-server/admin/conversations/${userID}/${convID}/ws?token=${encodeURIComponent(token)}`;

    const ws = new WebSocket(url);
    watchWsRef.current = ws;

    ws.onmessage = (event) => {
      if (watchWsRef.current !== ws) return;
      try {
        const data = JSON.parse(event.data);

        // Conversation ended signal from the hub.
        if (data.type === "ended") {
          setConvEnded(true);
          return;
        }

        if (data.sender && data.message) {
          setMessages((prev) => [
            ...prev,
            {
              id: data.id ?? Date.now(),
              sender: data.sender,
              message: data.message,
              timestamp: data.timestamp,
            },
          ]);
        }
      } catch (e) {
        console.error("admin watch WS parse error", e);
      }
    };

    ws.onerror = () => {
      console.error("admin watch WS error");
    };

    ws.onclose = () => {
      if (watchWsRef.current === ws) watchWsRef.current = null;
    };
  };

  // Close the watch WS when the component unmounts.
  useEffect(() => () => closeWatchWs(), []);

  // Single persistent WebSocket that subscribes to ALL user messages across
  // all conversations.  When a message arrives for a conv that isn't currently
  // selected, we mark it unread in the sidebar.
  useEffect(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const token = UserService.getToken();
    const url = `${protocol}//${window.location.host}/pac-go-server/admin/watch?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(url);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (!data.user_id || !data.conversation_id) return;
        const sel = selectedRef.current;
        const isSelected =
          sel?.userID === data.user_id && sel?.convID === data.conversation_id;
        if (!isSelected) {
          const key = convKey(data.user_id, data.conversation_id);
          setUnreadCounts((prev) => ({ ...prev, [key]: (prev[key] ?? 0) + 1 }));
        }
      } catch (e) {
        console.error("admin watch-all WS parse error", e);
      }
    };

    ws.onerror = () => console.error("admin watch-all WS error");

    return () => {
      ws.onmessage = null;
      ws.onerror = null;
      ws.onclose = null;
      ws.close(1000, "unmount");
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const selectConversation = async (userID, convID, alreadyEnded = false, username = "") => {
    const key = convKey(userID, convID);
    setSelected({ userID, convID, username });
    setMessages([]);
    setReplyText("");
    setSuccess("");
    setConvEnded(alreadyEnded); // restore persisted ended state immediately
    // Clear unread count for this conv.
    setUnreadCounts((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
    setMsgLoading(true);
    try {
      const res = await api.get(`/pac-go-server/admin/conversations/${userID}/${convID}/messages`);
      const msgs = res.data.messages ?? [];
      setMessages(msgs);
    } catch (err) {
      setError("Failed to load messages: " + (err.response?.data?.error ?? err.message));
    } finally {
      setMsgLoading(false);
    }
    // Open the live-watch WebSocket after history is loaded.
    openWatchWs(userID, convID);
  };

  const sendReply = async () => {
    if (!replyText.trim() || !selected || convEnded) return;
    setSuccess("");
    setError("");
    const text = replyText;
    setReplyText("");
    try {
      await api.post(
        `/pac-go-server/admin/conversations/${selected.userID}/${selected.convID}/reply`,
        { message: text }
      );
      setSuccess("Reply sent.");
      // Optimistically append the admin message so it appears immediately.
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now(),
          sender: "admin",
          message: text,
          timestamp: new Date().toISOString(),
        },
      ]);
    } catch (err) {
      setReplyText(text); // restore on failure
      setError("Failed to send reply: " + (err.response?.data?.error ?? err.message));
    }
  };

  // onKeyDown for the reply textarea: Enter submits, Shift+Enter adds a newline.
  const onReplyKeyDown = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendReply();
    }
  };

  const endConversation = async () => {
    if (!selected || convEnded) return;
    setError("");
    try {
      await api.post(
        `/pac-go-server/admin/conversations/${selected.userID}/${selected.convID}/end`
      );
      setConvEnded(true);
      setSuccess("Conversation ended.");
      // Refresh the conversation list so the new conv appears.
      loadConversations();
    } catch (err) {
      setError("Failed to end conversation: " + (err.response?.data?.error ?? err.message));
    }
  };

  const formatTime = (ts) =>
    new Date(ts).toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: true,
    });

  return (
    <div className="chat-support-container chat-support-container--admin">
      {/* Left panel — conversation list */}
      <div className="conversations-panel">
        <div className="conversations-header">
          <h3>Conversations</h3>
          <Button
            kind="ghost"
            size="sm"
            renderIcon={Renew}
            iconDescription="Refresh"
            onClick={loadConversations}
            hasIconOnly
          />
        </div>

        {loading && <DataTableSkeleton columnCount={1} rowCount={4} />}

        {!loading && conversations.length === 0 && (
          <p className="empty-state-text">No conversations yet</p>
        )}

        <div className="conversations-list">
          {conversations.map((conv) => {
            const isActive =
              selected?.userID === conv.user_id &&
              selected?.convID === conv.conversation_id;
            const unreadCount = unreadCounts[convKey(conv.user_id, conv.conversation_id)] ?? 0;
            return (
              <div
                key={`${conv.user_id}-${conv.conversation_id}`}
                className={`conversation-item${isActive ? " active" : ""}${conv.ended ? " conversation-item--ended" : ""}`}
                onClick={() => selectConversation(conv.user_id, conv.conversation_id, conv.ended, conv.username)}
              >
                <div className="conversation-avatar">
                  <div className={`avatar-circle${conv.ended ? " avatar-circle--ended" : ""}`}>
                    {(conv.username || conv.user_id)?.[0]?.toUpperCase() ?? "U"}
                  </div>
                </div>
                <div className="conversation-details">
                  <div className="conversation-id">
                    {convLabel(conv.username, conv.user_id, conv.conversation_id)}
                  </div>
                  <div className="conversation-status">
                    {conv.ended
                      ? <span className="conv-ended-badge">Ended</span>
                      : <span className="status-connected">Active</span>}
                  </div>
                  <div className="conversation-unread">
                    {unreadCount > 0
                      ? <><span className="unread-dot" />{unreadCount} unread</>
                      : null}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Right panel — message thread + reply */}
      <div className="chat-panel">
        {error && (
          <InlineNotification
            kind="error"
            title="Error"
            subtitle={error}
            onCloseButtonClick={() => setError("")}
          />
        )}
        {success && (
          <InlineNotification
            kind="success"
            title="Success"
            subtitle={success}
            onCloseButtonClick={() => setSuccess("")}
          />
        )}

        {!selected ? (
          <div className="empty-state">
            <p>Select a conversation on the left to view messages</p>
          </div>
        ) : (
          <>
            <div className="chat-header">
              <h2>{convLabel(selected.username, selected.userID, selected.convID)}</h2>
              <div className="chat-status">
                <span className="status-label" style={{ fontWeight: 600, marginRight: 8 }}>
                  User: {selected.username || selected.userID}
                </span>
                {!convEnded && (
                  <button
                    className="end-conv-button"
                    onClick={endConversation}
                    title="End this conversation"
                  >
                    End Conversation
                  </button>
                )}
                {convEnded && (
                  <span className="conv-ended-badge">Ended</span>
                )}
              </div>
            </div>

            <div className="chat-messages">
              {msgLoading && <DataTableSkeleton columnCount={1} rowCount={3} />}
              {!msgLoading && messages.length === 0 && (
                <div className="empty-state">
                  <p>No messages in this conversation</p>
                </div>
              )}
              {!msgLoading &&
                messages.map((msg, idx) => (
                  <div
                    key={msg.id ?? idx}
                    className={`message ${msg.sender === "user" ? "message-user" : "message-admin"}`}
                  >
                    <div className="message-content">
                      <div className="message-header">
                        <span className="message-sender">
                          [{msg.sender === "user" ? "User" : "Admin"},{" "}
                          {formatTime(msg.timestamp)}]:
                        </span>
                      </div>
                      <div className="message-text">{msg.message}</div>
                    </div>
                  </div>
                ))}
              <div ref={messagesEndRef} />
            </div>

            <div className="chat-input-container">
              <div className="chat-input-wrapper admin-reply-wrapper">
                <TextArea
                  className="chat-textarea"
                  placeholder={convEnded ? "Conversation ended" : "Type your reply… (Enter to send, Shift+Enter for new line)"}
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  onKeyDown={onReplyKeyDown}
                  rows={2}
                  labelText=""
                  disabled={convEnded}
                />
                <button
                  className="send-button"
                  onClick={sendReply}
                  disabled={!replyText.trim() || convEnded}
                >
                  <Send size={20} />
                  <span>Reply</span>
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default ChatAdmin;
