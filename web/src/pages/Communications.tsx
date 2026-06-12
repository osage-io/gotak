/**
 * Communications Page - Modern Tactical Chat & Messaging System
 * Responsive design with hamburger menu for groups/users
 */

import React, { useState, useEffect, useRef } from 'react';
import { wsService } from '../services/websocketService';
import { ChatMessage, ChatRoom } from '../services/apiClient';
import { Icon } from '../components/ui/Icon';
import AIChat from '../components/AIChat';
import './Communications.css';

// --- Vault transit encryption (DEMO ONLY) ---------------------------------
// For the demo we talk to the local Vault directly with the dev root token.
// Do NOT ship a root token in frontend code outside a throwaway demo.
const VAULT_ADDR = (typeof window !== 'undefined' && (window as any).GOTAK_CONFIG?.vaultUrl) || 'http://127.0.0.1:8200';
const VAULT_TOKEN = 'root';
const DEFAULT_TRANSIT_KEY = 'gotak-comms';

// Each channel gets its own transit key, so a channel is "Encrypted" exactly
// when this key exists in Vault. The AI Intel Officer is treated as its own
// channel so its comms can be secured the same way.
const AI_OFFICER_ID = 'ai-officer';
const channelKeyName = (roomId: string) => `gotak-comms-${roomId}`;

// base64-encode a UTF-8 string the way Vault expects for transit plaintext
function toBase64Utf8(str: string): string {
  return btoa(unescape(encodeURIComponent(str)));
}

// True if the named transit key exists in Vault (GET returns 200).
async function vaultKeyExists(addr: string, token: string, keyName: string): Promise<boolean> {
  try {
    const res = await fetch(`${addr}/v1/transit/keys/${encodeURIComponent(keyName)}`, {
      headers: { 'X-Vault-Token': token },
    });
    return res.ok;
  } catch {
    return false;
  }
}

// Create the named transit key (no-op server-side if it already exists).
async function vaultCreateKey(addr: string, token: string, keyName: string): Promise<void> {
  const res = await fetch(`${addr}/v1/transit/keys/${encodeURIComponent(keyName)}`, {
    method: 'POST',
    headers: { 'X-Vault-Token': token, 'Content-Type': 'application/json' },
    body: JSON.stringify({ type: 'aes256-gcm96' }),
  });
  if (!res.ok) throw new Error(`Vault create-key failed (HTTP ${res.status})`);
}

// Encrypt plaintext via Vault's transit engine, returning the "vault:v1:..."
// ciphertext. Throws on any non-2xx response.
async function vaultEncrypt(
  plaintext: string,
  keyName: string,
  addr: string = VAULT_ADDR,
  token: string = VAULT_TOKEN,
): Promise<string> {
  const res = await fetch(`${addr}/v1/transit/encrypt/${encodeURIComponent(keyName)}`, {
    method: 'POST',
    headers: {
      'X-Vault-Token': token,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ plaintext: toBase64Utf8(plaintext) }),
  });
  if (!res.ok) {
    throw new Error(`Vault encrypt failed (HTTP ${res.status})`);
  }
  const data = await res.json();
  return data?.data?.ciphertext as string;
}

interface ChatState {
  rooms: ChatRoom[];
  activeRoomId: string | null;
  messages: Map<string, ChatMessage[]>;
  typingUsers: Map<string, Set<string>>;
  onlineUsers: Set<string>;
  encryptionEnabled: boolean;
  showAIChat: boolean;
}

const Communications: React.FC = () => {
  const [chatState, setChatState] = useState<ChatState>({
    rooms: [],
    activeRoomId: 'general',
    messages: new Map(),
    typingUsers: new Map(),
    onlineUsers: new Set(),
    encryptionEnabled: false, // This would be set based on Vault integration status
    showAIChat: false,
  });

  const [messageInput, setMessageInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [encrypting, setEncrypting] = useState(false);
  const [encryptError, setEncryptError] = useState<string | null>(null);

  // Per-channel encryption: roomIds whose Vault transit key exists.
  const [encryptedRooms, setEncryptedRooms] = useState<Set<string>>(new Set());
  const [showVaultModal, setShowVaultModal] = useState(false);
  const [vaultBusy, setVaultBusy] = useState(false);
  const [vaultError, setVaultError] = useState<string | null>(null);
  const [vaultForm, setVaultForm] = useState({
    addr: VAULT_ADDR,
    token: VAULT_TOKEN,
    keyName: '',
  });
  // The channel the encryption modal is acting on (a room, or the AI officer).
  const [vaultTarget, setVaultTarget] = useState<{ id: string; label: string } | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [showNewRoomDialog, setShowNewRoomDialog] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll to bottom of messages
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatState.messages, chatState.activeRoomId]);

  // Check URL parameters for AI chat
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('ai') === 'true') {
      setChatState(prev => ({ ...prev, showAIChat: true, activeRoomId: null }));
    }
  }, []);

  // Initialize chat rooms and WebSocket listeners
  useEffect(() => {
    // Mock initial rooms - in real app, fetch from API
    const initialRooms: ChatRoom[] = [
      {
        id: 'general',
        name: 'General',
        type: 'public',
        members: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 'tactical',
        name: 'Tactical Ops',
        type: 'private',
        members: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 'intel',
        name: 'Intelligence',
        type: 'private',
        members: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 'alpha-team',
        name: 'Alpha Team',
        type: 'team',
        members: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];

    // Mock some messages
    const mockMessages = new Map<string, ChatMessage[]>();
    mockMessages.set('general', [
      {
        id: '1',
        roomId: 'general',
        senderId: 'user1',
        senderCallsign: 'ALPHA-1',
        content: 'System check complete. All units operational.',
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        priority: 'normal',
        classification: 'UNCLASSIFIED',
      },
      {
        id: '2',
        roomId: 'general',
        senderId: 'user2',
        senderCallsign: 'BRAVO-2',
        content: 'Roger that. Standing by for orders.',
        timestamp: new Date(Date.now() - 1800000).toISOString(),
        priority: 'normal',
        classification: 'UNCLASSIFIED',
      },
    ]);

    setChatState(prev => ({
      ...prev,
      rooms: initialRooms,
      activeRoomId: 'general',
      messages: mockMessages,
      onlineUsers: new Set(['ALPHA-1', 'BRAVO-2', 'CHARLIE-3', 'DELTA-4']),
    }));

    // Set up WebSocket listeners
    const unsubscribeChatMessage = wsService.onChatMessage(({ message, action }) => {
      if (action === 'new') {
        setChatState(prev => {
          const newMessages = new Map(prev.messages);
          const roomMessages = newMessages.get(message.roomId) || [];
          newMessages.set(message.roomId, [...roomMessages, message]);
          return { ...prev, messages: newMessages };
        });
      }
    });

    const unsubscribeUserTyping = wsService.onUserTyping(({ roomId, userId, username, typing }) => {
      setChatState(prev => {
        const newTypingUsers = new Map(prev.typingUsers);
        const roomTypingUsers = newTypingUsers.get(roomId) || new Set();
        
        if (typing) {
          roomTypingUsers.add(`${username}`);
        } else {
          roomTypingUsers.delete(`${username}`);
        }
        
        newTypingUsers.set(roomId, roomTypingUsers);
        return { ...prev, typingUsers: newTypingUsers };
      });
    });

    return () => {
      unsubscribeChatMessage();
      unsubscribeUserTyping();
    };
  }, []);

  // Handle typing indicator
  useEffect(() => {
    if (isTyping && chatState.activeRoomId) {
      wsService.setTyping(chatState.activeRoomId, true);
      
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      
      typingTimeoutRef.current = setTimeout(() => {
        setIsTyping(false);
        if (chatState.activeRoomId) {
          wsService.setTyping(chatState.activeRoomId, false);
        }
      }, 2000);
    }
  }, [isTyping, chatState.activeRoomId]);

  // Send message
  const sendMessage = () => {
    if (!messageInput.trim() || !chatState.activeRoomId) return;

    const newMessage: ChatMessage = {
      id: `msg-${Date.now()}`,
      roomId: chatState.activeRoomId,
      senderId: 'current-user',
      senderCallsign: 'YOU',
      content: messageInput.trim(),
      timestamp: new Date().toISOString(),
      priority: 'normal',
      classification: 'UNCLASSIFIED',
    };

    // Add message locally
    setChatState(prev => {
      const newMessages = new Map(prev.messages);
      const roomMessages = newMessages.get(chatState.activeRoomId!) || [];
      newMessages.set(chatState.activeRoomId!, [...roomMessages, newMessage]);
      return { ...prev, messages: newMessages };
    });

    // Send via WebSocket
    wsService.sendChatMessage(chatState.activeRoomId, messageInput.trim(), {
      messageType: 'text',
      priority: 'normal',
      classification: 'UNCLASSIFIED',
    });

    setMessageInput('');
    setIsTyping(false);
  };

  // Encrypt the current input via Vault transit, then send it as a new,
  // encrypted message (shows a green "Encrypted" badge).
  const sendEncryptedMessage = async () => {
    if (!messageInput.trim() || !chatState.activeRoomId || encrypting) return;
    setEncrypting(true);
    setEncryptError(null);
    try {
      // Use this channel's key if encryption is configured for it; otherwise
      // fall back to the shared demo key.
      const roomId = chatState.activeRoomId;
      const keyName = encryptedRooms.has(roomId) ? channelKeyName(roomId) : DEFAULT_TRANSIT_KEY;
      const ciphertext = await vaultEncrypt(messageInput.trim(), keyName);

      const newMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        roomId: chatState.activeRoomId,
        senderId: 'current-user',
        senderCallsign: 'YOU',
        content: ciphertext,
        timestamp: new Date().toISOString(),
        priority: 'normal',
        classification: 'UNCLASSIFIED',
        encrypted: true,
        transitKey: keyName,
      } as ChatMessage;

      setChatState(prev => {
        const newMessages = new Map(prev.messages);
        const roomMessages = newMessages.get(chatState.activeRoomId!) || [];
        newMessages.set(chatState.activeRoomId!, [...roomMessages, newMessage]);
        return { ...prev, messages: newMessages };
      });

      // Send the ciphertext over the wire — never the plaintext.
      wsService.sendChatMessage(chatState.activeRoomId, ciphertext, {
        messageType: 'text',
        priority: 'normal',
        classification: 'UNCLASSIFIED',
      });

      setMessageInput('');
      setIsTyping(false);
    } catch (err) {
      setEncryptError(err instanceof Error ? err.message : 'Encryption failed');
    } finally {
      setEncrypting(false);
    }
  };

  // When the current target (active room, or the AI officer) changes, check
  // Vault for its transit key so the indicator reflects the real state.
  useEffect(() => {
    const targetId = chatState.showAIChat ? AI_OFFICER_ID : chatState.activeRoomId;
    if (!targetId) return;
    let cancelled = false;
    vaultKeyExists(VAULT_ADDR, VAULT_TOKEN, channelKeyName(targetId)).then(exists => {
      if (cancelled) return;
      setEncryptedRooms(prev => {
        const next = new Set(prev);
        if (exists) next.add(targetId); else next.delete(targetId);
        return next;
      });
    });
    return () => { cancelled = true; };
  }, [chatState.activeRoomId, chatState.showAIChat]);

  // Open the "Configure Vault Encryption" modal for the current target.
  const openVaultModal = () => {
    const targetId = chatState.showAIChat ? AI_OFFICER_ID : chatState.activeRoomId;
    if (!targetId) return;
    const label = chatState.showAIChat
      ? 'AI Intel Officer'
      : (chatState.rooms.find(r => r.id === targetId)?.name ?? targetId);
    setVaultTarget({ id: targetId, label });
    setVaultForm({ addr: VAULT_ADDR, token: VAULT_TOKEN, keyName: channelKeyName(targetId) });
    setVaultError(null);
    setShowVaultModal(true);
  };

  // Create the transit key in Vault, which "turns on" encryption for the target.
  const enableChannelEncryption = async () => {
    const targetId = vaultTarget?.id;
    if (!targetId || vaultBusy) return;
    setVaultBusy(true);
    setVaultError(null);
    try {
      const exists = await vaultKeyExists(vaultForm.addr, vaultForm.token, vaultForm.keyName);
      if (!exists) {
        await vaultCreateKey(vaultForm.addr, vaultForm.token, vaultForm.keyName);
      }
      setEncryptedRooms(prev => new Set(prev).add(targetId));
      setShowVaultModal(false);
    } catch (err) {
      setVaultError(err instanceof Error ? err.message : 'Failed to enable encryption');
    } finally {
      setVaultBusy(false);
    }
  };

  // Handle message input changes
  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setMessageInput(e.target.value);
    if (!isTyping && e.target.value.trim()) {
      setIsTyping(true);
    }
  };

  // Handle Enter key
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // Switch room
  const switchRoom = (roomId: string) => {
    setChatState(prev => ({ ...prev, activeRoomId: roomId, showAIChat: false }));
    setMenuOpen(false);
    // Focus input after switching
    setTimeout(() => inputRef.current?.focus(), 100);
  };

  // Open AI Chat
  const openAIChat = () => {
    setChatState(prev => ({ ...prev, showAIChat: true, activeRoomId: null }));
    setMenuOpen(false);
  };

  // Create new room
  const createNewRoom = () => {
    if (!newRoomName.trim()) return;

    const newRoom: ChatRoom = {
      id: `room-${Date.now()}`,
      name: newRoomName.trim(),
      type: 'private',
      members: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    setChatState(prev => ({
      ...prev,
      rooms: [...prev.rooms, newRoom],
      activeRoomId: newRoom.id,
      messages: new Map(prev.messages.set(newRoom.id, [])),
    }));

    setNewRoomName('');
    setShowNewRoomDialog(false);
    setMenuOpen(false);
  };

  // Filter rooms and users based on search
  const filteredRooms = chatState.rooms.filter(room =>
    room.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredUsers = Array.from(chatState.onlineUsers).filter(user =>
    user.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const activeRoom = chatState.rooms.find(r => r.id === chatState.activeRoomId);
  const isActiveEncrypted = chatState.activeRoomId ? encryptedRooms.has(chatState.activeRoomId) : false;
  // Encryption "target" = the AI officer when its page is open, else the room.
  const encryptionTargetId = chatState.showAIChat ? AI_OFFICER_ID : chatState.activeRoomId;
  const encryptionTargetLabel = chatState.showAIChat ? 'AI Intel Officer' : (activeRoom?.name ?? '');
  const isTargetEncrypted = encryptionTargetId ? encryptedRooms.has(encryptionTargetId) : false;
  const activeMessages = chatState.activeRoomId ? 
    chatState.messages.get(chatState.activeRoomId) || [] : [];
  const typingInActiveRoom = chatState.activeRoomId ?
    chatState.typingUsers.get(chatState.activeRoomId) || new Set() : new Set();

  return (
    <div className="comms-container">
      {/* Integrated Sidebar */}
      <aside className={`comms-sidebar ${menuOpen ? 'collapsed' : ''}`}>
        <div className="sidebar-header">
          <h2 className="sidebar-title">COMMS</h2>
          <button 
            className="sidebar-toggle"
            onClick={() => setMenuOpen(!menuOpen)}
            aria-label="Toggle Sidebar"
            title={menuOpen ? 'Expand Sidebar' : 'Collapse Sidebar'}
          >
            {menuOpen ? '→' : '←'}
          </button>
        </div>

        {!menuOpen && (
          <>
            {/* Search Bar */}
            <div className="sidebar-search">
              <input
                type="text"
                placeholder="Search..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
              <span className="search-icon">
                <Icon name="search" size={16} color="var(--color-text-muted)" />
              </span>
            </div>

            {/* Rooms Section */}
            <div className="sidebar-section">
              <div className="section-header">
                <h3>Channels</h3>
                <button 
                  className="add-btn"
                  onClick={() => setShowNewRoomDialog(true)}
                  title="Create Channel"
                >
                  +
                </button>
              </div>
              <div className="rooms-list">
                {filteredRooms.map(room => (
                  <button
                    key={room.id}
                    className={`room-item ${room.id === chatState.activeRoomId ? 'active' : ''}`}
                    onClick={() => switchRoom(room.id)}
                  >
                    <span className="room-type-icon">
                      <Icon 
                        name={room.type === 'private' ? 'lock' : room.type === 'team' ? 'users' : 'chat'} 
                        size={14} 
                        color="var(--color-text-secondary)" 
                      />
                    </span>
                    <span className="room-name">{room.name}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* AI Intelligence Section */}
            <div className="sidebar-section ai-section">
              <div className="section-header">
                <h3>AI Intelligence</h3>
              </div>
              <div className="ai-list">
                <button
                  className={`ai-item ${chatState.showAIChat ? 'active' : ''}`}
                  onClick={openAIChat}
                >
                  <span className="ai-status-icon">
                    <Icon name="bot" size={14} color="#00ff41" />
                  </span>
                  <span className="ai-name">AI Intel Officer</span>
                  <span className="ai-badge">READY</span>
                </button>
              </div>
            </div>

            {/* Online Users Section */}
            <div className="sidebar-section">
              <div className="section-header">
                <h3>Online — {filteredUsers.length}</h3>
              </div>
              <div className="users-list">
                {filteredUsers.map(user => (
                  <div key={user} className="user-item">
                    <span className="user-status"></span>
                    <span className="user-name">{user}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Encryption Status (active room, or the AI officer) */}
            <div className="sidebar-section encryption-section">
              <div className="encryption-status-card">
                <div className={`encryption-indicator ${isTargetEncrypted ? 'encrypted' : 'unencrypted'}`}>
                  <span className="encryption-icon">
                    <Icon
                      name={isTargetEncrypted ? 'lock' : 'unlock'}
                      size={20}
                      color={isTargetEncrypted ? 'var(--color-success)' : 'var(--color-warning)'}
                    />
                  </span>
                  <div className="encryption-info">
                    <span className="encryption-label">
                      Encryption{encryptionTargetLabel ? ` — ${encryptionTargetLabel}` : ''}
                    </span>
                    <span className="encryption-state">
                      {isTargetEncrypted ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                </div>
                <div className="encryption-details">
                  {isTargetEncrypted ? (
                    <span className="encryption-text">Messages are end-to-end encrypted via Vault Transit</span>
                  ) : (
                    <span className="encryption-text warning">Messages are not encrypted</span>
                  )}
                </div>
                {!isTargetEncrypted && encryptionTargetId && (
                  <button className="encryption-configure-btn" onClick={openVaultModal}>
                    <Icon name="lock" size={14} />
                    Configure Encryption
                  </button>
                )}
              </div>
            </div>
          </>
        )}
      </aside>

      {/* Main Chat Area */}
      <main className="chat-main">
        {chatState.showAIChat ? (
          /* AI Chat Interface */
          <AIChat onClose={() => setChatState(prev => ({ ...prev, showAIChat: false, activeRoomId: 'general' }))} />
        ) : activeRoom ? (
          <>
            {/* Chat Header */}
            <header className="chat-header">
              <div className="channel-info">
                <span className="channel-icon">
                  <Icon 
                    name={activeRoom.type === 'private' ? 'lock' : activeRoom.type === 'team' ? 'users' : 'chat'} 
                    size={18} 
                    color="var(--color-accent)" 
                  />
                </span>
                <h2 className="channel-name">{activeRoom.name}</h2>
              </div>
              <div className="header-actions">
                {isActiveEncrypted ? (
                  <div className="encryption-badge encrypted" title={`Vault transit key: ${channelKeyName(activeRoom.id)}`}>
                    <span className="lock-icon">
                      <Icon name="lock" size={14} color="var(--color-success)" />
                    </span>
                    <span className="encryption-text">Encrypted</span>
                  </div>
                ) : (
                  <button
                    className="encryption-badge configure-encryption"
                    onClick={openVaultModal}
                    title="Configure Vault encryption for this channel"
                  >
                    <span className="lock-icon">
                      <Icon name="unlock" size={14} color="var(--color-warning)" />
                    </span>
                    <span className="encryption-text">Configure Encryption</span>
                  </button>
                )}
                <span className="member-count">
                  <span className="online-dot"></span>
                  {chatState.onlineUsers.size} Online
                </span>
              </div>
            </header>

            {/* Messages Container */}
            <div className="messages-container">
              <div className="messages-scroll">
                {activeMessages.length === 0 ? (
                  <div className="no-messages">
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                ) : (
                  activeMessages.map(message => (
                    <div 
                      key={message.id}
                      className={`message ${message.senderId === 'current-user' ? 'own' : ''}`}
                    >
                      <div className="message-header">
                        <span className="sender">{message.senderCallsign}</span>
                        {(message as ChatMessage & { encrypted?: boolean; transitKey?: string }).encrypted && (
                          <span
                            className="encrypted-badge"
                            title={`Encrypted with Vault transit key "${(message as ChatMessage & { transitKey?: string }).transitKey || DEFAULT_TRANSIT_KEY}"`}
                          >
                            <Icon name="lock" size={11} />
                            Encrypted
                          </span>
                        )}
                        <span className="time">
                          {new Date(message.timestamp).toLocaleTimeString([], { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })}
                        </span>
                      </div>
                      <div className={`message-body${(message as ChatMessage & { encrypted?: boolean }).encrypted ? ' message-body-encrypted' : ''}`}>
                        {message.content}
                      </div>
                    </div>
                  ))
                )}
                
                {/* Typing Indicator */}
                {typingInActiveRoom.size > 0 && (
                  <div className="typing-indicator">
                    <div className="typing-dots">
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                    <span className="typing-text">
                      {Array.from(typingInActiveRoom).join(', ')} typing...
                    </span>
                  </div>
                )}
                
                <div ref={messagesEndRef} />
              </div>
            </div>

            {/* Message Input Area */}
            <div className="input-area">
              <div className="input-container">
                <textarea
                  ref={inputRef}
                  value={messageInput}
                  onChange={handleInputChange}
                  onKeyDown={handleKeyPress}
                  placeholder="Type a message..."
                  className="message-input"
                  rows={1}
                />
                <button
                  className="encrypt-btn"
                  onClick={sendEncryptedMessage}
                  disabled={!messageInput.trim() || encrypting}
                  aria-label="Encrypt and send message"
                  title={`Encrypt with Vault transit key "${isActiveEncrypted && activeRoom ? channelKeyName(activeRoom.id) : DEFAULT_TRANSIT_KEY}"`}
                >
                  <Icon name="lock" size={16} />
                  <span>{encrypting ? 'Encrypting…' : 'Encrypt'}</span>
                </button>
                <button
                  className="send-btn"
                  onClick={sendMessage}
                  disabled={!messageInput.trim()}
                  aria-label="Send Message"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M22 2L11 13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </button>
              </div>
              {encryptError && (
                <div className="encrypt-error" role="alert">{encryptError}</div>
              )}
            </div>
          </>
        ) : (
          <div className="no-room">
            <p>Select a room to start messaging</p>
          </div>
        )}
      </main>

      {/* Configure Vault Encryption Dialog */}
      {showVaultModal && vaultTarget && (
        <div className="modal-overlay" onClick={() => !vaultBusy && setShowVaultModal(false)}>
          <div className="modal-dialog vault-encrypt-dialog" onClick={e => e.stopPropagation()}>
            <h3>
              <Icon name="lock" size={18} color="var(--color-accent)" />
              Configure Encryption — {vaultTarget.label}
            </h3>
            <p className="vault-dialog-desc">
              Creates a dedicated Vault transit key for this channel. Once the key
              exists, the channel is marked <strong>Encrypted</strong>.
            </p>

            <div className="vault-field">
              <label>Vault Address</label>
              <input
                type="text"
                className="modal-input"
                value={vaultForm.addr}
                onChange={e => setVaultForm(f => ({ ...f, addr: e.target.value }))}
              />
            </div>
            <div className="vault-field">
              <label>Vault Token</label>
              <input
                type="password"
                className="modal-input"
                value={vaultForm.token}
                onChange={e => setVaultForm(f => ({ ...f, token: e.target.value }))}
              />
            </div>
            <div className="vault-field">
              <label>Transit Key Name</label>
              <input
                type="text"
                className="modal-input"
                value={vaultForm.keyName}
                onChange={e => setVaultForm(f => ({ ...f, keyName: e.target.value }))}
              />
              <p className="vault-field-help">This key is created in Vault's transit engine.</p>
            </div>

            {vaultError && <div className="vault-dialog-error" role="alert">{vaultError}</div>}

            <div className="modal-actions">
              <button className="btn-cancel" onClick={() => setShowVaultModal(false)} disabled={vaultBusy}>
                Cancel
              </button>
              <button
                className="btn-create"
                onClick={enableChannelEncryption}
                disabled={vaultBusy || !vaultForm.addr.trim() || !vaultForm.token.trim() || !vaultForm.keyName.trim()}
              >
                {vaultBusy ? 'Enabling…' : 'Enable Encryption'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* New Room Dialog */}
      {showNewRoomDialog && (
        <div className="modal-overlay" onClick={() => setShowNewRoomDialog(false)}>
          <div className="modal-dialog" onClick={e => e.stopPropagation()}>
            <h3>Create New Room</h3>
            <input
              type="text"
              value={newRoomName}
              onChange={(e) => setNewRoomName(e.target.value)}
              placeholder="Enter room name..."
              className="modal-input"
              autoFocus
              onKeyPress={(e) => e.key === 'Enter' && createNewRoom()}
            />
            <div className="modal-actions">
              <button 
                className="btn-cancel"
                onClick={() => setShowNewRoomDialog(false)}
              >
                Cancel
              </button>
              <button 
                className="btn-create"
                onClick={createNewRoom}
                disabled={!newRoomName.trim()}
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Communications;
