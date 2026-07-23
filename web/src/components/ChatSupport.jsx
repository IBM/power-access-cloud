import React, { useState, useEffect, useRef, useCallback } from "react";
import { Send, Add } from "@carbon/icons-react";
import UserService from "../services/UserService";
import axios from "axios";
import "../styles/chat-support.scss";

// Axios instance that attaches the Bearer token to every request.
const _axios = axios.create();
_axios.interceptors.request.use((config) => {
  if (UserService.isLoggedIn()) {
    const cb = () => {
      config.headers.Authorization = `Bearer ${UserService.getToken()}`;
      return Promise.resolve(config);
    };
    return UserService.updateToken(cb);
  }
});

const ChatSupport = () => {
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState("");
  const [conversationId, setConversationId] = useState(null);   // numeric ID from server
  const [connectionStatus, setConnectionStatus] = useState("connecting");
  const [wsOpen, setWsOpen] = useState(false);
  const [ended, setEnded] = useState(false);  // true after user ends conv, before new one starts
  // pastConversations: list of {id, messageCount, messages|null} — messages loaded lazily on click
  const [pastConversations, setPastConversations] = useState([]);
  // selectedPastId: which past conv is selected in the sidebar (null = current active conv)
  const [selectedPastId, setSelectedPastId] = useState(null);
  // unreadCounts: { [convId]: number } — admin replies received while that conv is not in view
  const [unreadCounts, setUnreadCounts] = useState({});
  const wsRef = useRef(null);
  const reconnectTimer = useRef(null);
  const messagesEndRef = useRef(null);
  // Flipped to false on intentional unmount so the onclose handler
  // does not schedule a reconnect after the component is gone.
  const shouldReconnect = useRef(true);
  // Set to true by startNewConversation so the next hello frame does not
  // restore a persisted ended state for the same conv ID.
  const startingNewConv = useRef(false);
  // Holds the fetched conversation summaries so the hello WS handler can
  // check ended status regardless of which resolves first (fetch vs WS).
  const fetchedConvsRef = useRef([]);

  const username = UserService.getUsername() ?? "user";

  // Derived display label: "username-conv-N"
  const convLabel = conversationId != null ? `${username}-conv-${conversationId}` : null;

  // On mount: fetch the user's past conversations from the server so the sidebar
  // is populated even after a page navigation.
  useEffect(() => {
    _axios.get("/pac-go-server/conversations")
      .then((res) => {
        const convs = res.data.conversations ?? [];
        const mapped = convs.map((c) => ({
          id: c.conversation_id,
          messageCount: c.message_count,
          firstMessage: c.first_message || `Conversation ${c.conversation_id}`,
          ended: c.ended,
          messages: null,
        }));
        fetchedConvsRef.current = mapped;
        setPastConversations(mapped);
        // If the WS hello already arrived and set conversationId, apply ended now.
        // Use a functional update to read the latest conversationId.
        if (!startingNewConv.current) {
          setConversationId((currentId) => {
            if (currentId != null) {
              const entry = mapped.find((c) => c.id === currentId);
              if (entry?.ended) setTimeout(() => setEnded(true), 0);
            }
            return currentId;
          });
        }
      })
      .catch((err) => console.error("failed to load conversations", err));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    shouldReconnect.current = true;

    const buildWsUrl = () => {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const token = UserService.getToken();
      return `${protocol}//${window.location.host}/pac-go-server/chat?token=${encodeURIComponent(token)}`;
    };

    const connect = () => {
      // Close any existing socket before opening a new one.
      // Detach onclose first so the close doesn't schedule a reconnect.
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.onerror = null;
        wsRef.current.close(1000, "reconnecting");
      }

      const ws = new WebSocket(buildWsUrl());
      wsRef.current = ws;

      ws.onopen = () => {
        if (wsRef.current !== ws) return;
        setConnectionStatus("connected");
        setWsOpen(true);
      };

      ws.onmessage = (event) => {
        if (wsRef.current !== ws) return;
        try {
          const data = JSON.parse(event.data);

          // Hello frame: sent immediately on connect with conversation_id + history.
          if (data.type === "hello") {
            setConversationId(data.conversation_id);
            const isStartingNew = startingNewConv.current;
            startingNewConv.current = false; // consume the flag
            // Restore ended state from the persisted summaries on mount, or from
            // in-memory state if the user just ended it in this session.
            // Never restore ended when the user explicitly started a new conv.
            setEnded((prevEnded) => {
              if (isStartingNew) return false; // user started fresh — never ended
              if (prevEnded && data.conversation_id === conversationId) return true;
              // Check the already-fetched summaries ref (avoids the race where
              // the hello frame arrives before setPastConversations completes).
              const entry = fetchedConvsRef.current.find((c) => c.id === data.conversation_id);
              if (entry?.ended) {
                setTimeout(() => setEnded(true), 0);
              }
              return false;
            });
            const history = data.history ?? [];
            setMessages(
              history.map((m) => ({
                id: m.id,
                sender: m.sender === "user" ? "You" : "Admin",
                content: m.message,
                timestamp: new Date(m.timestamp),
              }))
            );
            return;
          }

          // Conversation ended (by user or by admin remote-end).
          // Keep the current messages visible in the left panel as a past conv.
          // Do NOT clear messages or assign a new conv ID yet — wait for the user
          // to explicitly click "Start new conversation".
          if (data.type === "ended") {
            setEnded(true);
            // Merge the in-memory messages into the pastConversations entry that was
            // pre-populated by the on-mount fetch (or add it if not present yet).
            setConversationId((currentId) => {
              setMessages((currentMessages) => {
                setPastConversations((prev) => {
                  const exists = prev.some((c) => c.id === currentId);
                  const firstMessage = currentMessages.find((m) => m.sender === "You")?.content
                    || `Conversation ${currentId}`;
                  if (exists) {
                    return prev.map((c) =>
                      c.id === currentId ? { ...c, ended: true, firstMessage, messages: currentMessages } : c
                    );
                  }
                  return [...prev, { id: currentId, ended: true, firstMessage, messageCount: currentMessages.length, messages: currentMessages }];
                });
                return currentMessages;
              });
              return currentId;
            });
            return;
          }

          // Echo after send: update conversationId in case it wasn't set yet.
          // Also clear the ended flag in fetchedConvsRef so navigation back
          // after sending a message on a previously-ended conv doesn't re-apply ended state.
          if (data.conversation_id) {
            setConversationId((prev) => prev ?? data.conversation_id);
            // Mark this conv as not-ended in the ref so the hello handler won't restore ended.
            fetchedConvsRef.current = fetchedConvsRef.current.map((c) =>
              c.id === data.conversation_id ? { ...c, ended: false } : c
            );
          }

          // Admin reply arriving mid-conversation.
          if (data.sender === "admin") {
            setMessages((prev) => [
              ...prev,
              {
                id: data.id ?? Date.now(),
                sender: "Admin",
                content: data.message,
                timestamp: data.timestamp ? new Date(data.timestamp) : new Date(),
              },
            ]);
            // If the user is currently viewing a past conv (not the active one),
            // increment the unread counter for the active conv.
            setSelectedPastId((currentPastId) => {
              if (currentPastId !== null) {
                setConversationId((cid) => {
                  if (cid != null) {
                    setUnreadCounts((prev) => ({ ...prev, [cid]: (prev[cid] ?? 0) + 1 }));
                  }
                  return cid;
                });
              }
              return currentPastId;
            });
          }

          if (data.error) {
            console.error("Server error:", data.error);
          }
        } catch (err) {
          console.error("Error parsing WS message", err);
        }
      };

      ws.onerror = () => {
        if (wsRef.current !== ws) return;
        setConnectionStatus("error");
      };

      ws.onclose = (event) => {
        if (wsRef.current !== ws) return;
        wsRef.current = null;
        setWsOpen(false);
        // 1006 = abnormal closure (upgrade rejected — 401 before WS handshake).
        // 1008 = policy violation.
        const authFailure = event.code === 1006 || event.code === 1008 || event.code === 4001;
        if (!shouldReconnect.current || authFailure) {
          setConnectionStatus("error");
          return;
        }
        setConnectionStatus("disconnected");
        reconnectTimer.current = setTimeout(connect, 3000);
      };
    };

    connect();

    return () => {
      shouldReconnect.current = false;
      clearTimeout(reconnectTimer.current);
      const ws = wsRef.current;
      wsRef.current = null;
      if (ws) {
        ws.onopen = null;
        ws.onclose = null;
        ws.onmessage = null;
        ws.onerror = null;
        ws.close(1000, "unmount");
      }
      setWsOpen(false);
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const formatTime = (date) =>
    date.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: true,
    });

  const statusLabel = {
    connected: "Active",
    connecting: "Connecting…",
    disconnected: "Reconnecting…",
    error: "Connection error — please refresh",
  };

  // startNewConversation: clear the view and allow the user to type.
  // The actual new conversation ID is assigned lazily by the server on the
  // first message — we just need to un-block the input here.
  const startNewConversation = useCallback(() => {
    startingNewConv.current = true; // tell the next hello frame to ignore persisted ended state
    setEnded(false);
    setMessages([]);
    setSelectedPastId(null);
    setConversationId(null); // clear so the header shows "Initializing…" until server echoes new ID
  }, []);

  const sendMessage = useCallback(() => {
    const ws = wsRef.current;
    if (!inputMessage.trim() || !ws || ws.readyState !== WebSocket.OPEN || ended) return;

    ws.send(JSON.stringify({ message: inputMessage }));

    setMessages((prev) => [
      ...prev,
      { id: Date.now(), sender: "You", content: inputMessage, timestamp: new Date() },
    ]);
    setInputMessage("");
  }, [inputMessage, ended]);

  const endConversation = useCallback(() => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN || ended) return;
    ws.send(JSON.stringify({ action: "end_conversation" }));
    // ended state is set when the server echoes back the "ended" frame,
    // not optimistically here, so the frame is the single source of truth.
  }, [ended]);

  const onKeyDown = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // selectPastConv: click handler that sets the selection and lazily fetches messages if not yet loaded.
  const selectPastConv = useCallback((convId) => {
    setSelectedPastId(convId);
    if (convId == null) return;
    setPastConversations((prev) => {
      const entry = prev.find((c) => c.id === convId);
      if (!entry || entry.messages !== null) return prev; // already loaded
      // Kick off the fetch; update the entry when it resolves.
      _axios.get(`/pac-go-server/conversations/${convId}/messages`)
        .then((res) => {
          const msgs = (res.data.messages ?? []).map((m) => ({
            id: m.id,
            sender: m.sender === "user" ? "You" : "Admin",
            content: m.message,
            timestamp: new Date(m.timestamp),
          }));
          setPastConversations((p) =>
            p.map((c) => c.id === convId ? { ...c, messages: msgs } : c)
          );
        })
        .catch((err) => console.error("failed to load messages for conv", convId, err));
      return prev; // no state change yet — fetch is async
    });
  }, []);

  // selectedPastConv: the past conversation object currently selected in the sidebar (or null)
  const selectedPastConv = selectedPastId != null
    ? pastConversations.find((c) => c.id === selectedPastId) ?? null
    : null;

  // displayedMessages / displayedConvId: what the right panel shows
  const displayedMessages = selectedPastConv ? (selectedPastConv.messages ?? []) : messages;
  const displayedConvId   = selectedPastConv ? selectedPastConv.id : conversationId;
  const displayedEnded    = selectedPastConv ? true : ended;
  const displayedLoading  = selectedPastConv && selectedPastConv.messages === null;

  return (
    <div className="chat-support-container">
      {/* Left panel */}
      <div className="conversations-panel">
        <div className="conversations-header">
          <h3>Conversations</h3>
        </div>
        <div className="conversations-list">
          {/* Current active/new conversation — always on top */}
          {!ended && (
            <div
              className={`conversation-item${!selectedPastId ? " active" : ""}`}
              onClick={() => {
                setSelectedPastId(null);
                // Clear unread for the active conv when clicking back to it.
                if (conversationId != null) {
                  setUnreadCounts((prev) => {
                    if (!prev[conversationId]) return prev;
                    const next = { ...prev };
                    delete next[conversationId];
                    return next;
                  });
                }
              }}
              style={{ cursor: "pointer" }}
            >
              <div className="conversation-avatar">
                <div className="avatar-circle">{username[0]?.toUpperCase() ?? "U"}</div>
              </div>
              <div className="conversation-details">
                <div className="conversation-id" title={messages.find((m) => m.sender === "You")?.content}>
                  {(() => {
                    const first = messages.find((m) => m.sender === "You")?.content;
                    if (!first) return "New conversation";
                    return first.length > 28 ? first.slice(0, 28) + "…" : first;
                  })()}
                </div>
                <div className={`conversation-status status-${connectionStatus}`}>
                  {statusLabel[connectionStatus]}
                </div>
                <div className="conversation-unread">
                  {conversationId != null && unreadCounts[conversationId] > 0
                    ? <><span className="unread-dot" />{unreadCounts[conversationId]} unread</>
                    : (messages.length === 0 ? "Start chatting" : null)}
                </div>
              </div>
            </div>
          )}

          {/* Past ended conversations — newest first (server returns DESC order) */}
          {pastConversations
            .filter((conv) => conv.id !== conversationId) // current active/ended conv shown separately above
            .map((conv) => (
            <div
              key={conv.id}
              className={`conversation-item conversation-item--ended${selectedPastId === conv.id ? " active" : ""}`}
              onClick={() => selectPastConv(conv.id)}
              style={{ cursor: "pointer" }}
            >
              <div className="conversation-avatar">
                <div className="avatar-circle avatar-circle--ended">{username[0]?.toUpperCase() ?? "U"}</div>
              </div>
              <div className="conversation-details">
                <div className="conversation-id" title={conv.firstMessage}>
                  {conv.firstMessage?.length > 28
                    ? conv.firstMessage.slice(0, 28) + "…"
                    : conv.firstMessage}
                </div>
                <div className="conversation-status">
                  <span className="conv-ended-badge">Ended</span>
                </div>
                <div className="conversation-unread" />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Right panel */}
      <div className="chat-panel">
        <div className="chat-header">
          <h2>
            {(() => {
              if (selectedPastConv) return selectedPastConv.firstMessage || `Conversation ${selectedPastConv.id}`;
              const first = messages.find((m) => m.sender === "You")?.content;
              return first || "New conversation";
            })()}
          </h2>
          <div className="chat-status">
            <span className={`status-${connectionStatus}`}>
              {displayedEnded ? "" : `Status: ${statusLabel[connectionStatus]}`}
            </span>
            {wsOpen && !ended && !selectedPastConv && (
              <button
                className="end-conv-button"
                onClick={endConversation}
                title="End this conversation"
              >
                End Conversation
              </button>
            )}
            {displayedEnded && <span className="conv-ended-badge">Ended</span>}
          </div>
        </div>

        <div className="chat-messages">
          {/* Spacer: grows to push a short message list to the bottom; collapses when list is long so scroll works */}
          <div style={{ flex: 1 }} />
          {displayedLoading ? (
            <div className="empty-state"><p>Loading…</p></div>
          ) : displayedMessages.length === 0 && !displayedEnded ? (
            <div className="empty-state">
              <p>Start your conversation by typing a message below</p>
            </div>
          ) : (
            displayedMessages.map((msg) => (
              <div
                key={msg.id}
                className={`message ${msg.sender === "You" ? "message-user" : "message-admin"}`}
              >
                <div className="message-content">
                  <div className="message-header">
                    <span className="message-sender">
                      [{msg.sender}, {formatTime(msg.timestamp)}]:
                    </span>
                  </div>
                  <div className="message-text">{msg.content}</div>
                </div>
              </div>
            ))
          )}
          {/* Ended divider — shown below the last message for ended convs */}
          {displayedEnded && (
            <div className="conv-ended-divider">
              <span>Conversation ended</span>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="chat-input-container">
          {ended ? (
            /* Conversation ended — show "Start new conversation" regardless of which past conv is selected */
            <div className="new-conv-prompt">
              <button
                className="new-conv-button"
                onClick={startNewConversation}
                disabled={!wsOpen}
              >
                <Add size={18} />
                Start new conversation
              </button>
            </div>
          ) : selectedPastConv ? null : (
            /* Active conv and no past conv selected — show the input bar */
            <div className="chat-input-wrapper">
              <input
                type="text"
                className="chat-input"
                placeholder="Type your message…"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyDown={onKeyDown}
                disabled={!wsOpen}
              />
              <button
                className="send-button"
                onClick={sendMessage}
                disabled={!inputMessage.trim() || !wsOpen}
              >
                <Send size={20} />
                <span>Send</span>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ChatSupport;
