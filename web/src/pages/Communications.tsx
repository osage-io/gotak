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

            {/* Encryption Status */}
            <div className="sidebar-section encryption-section">
              <div className="encryption-status-card">
                <div className={`encryption-indicator ${chatState.encryptionEnabled ? 'encrypted' : 'unencrypted'}`}>
                  <span className="encryption-icon">
                    <Icon 
                      name={chatState.encryptionEnabled ? 'lock' : 'unlock'} 
                      size={20} 
                      color={chatState.encryptionEnabled ? 'var(--color-success)' : 'var(--color-warning)'} 
                    />
                  </span>
                  <div className="encryption-info">
                    <span className="encryption-label">Encryption</span>
                    <span className="encryption-state">
                      {chatState.encryptionEnabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                </div>
                <div className="encryption-details">
                  {chatState.encryptionEnabled ? (
                    <span className="encryption-text">Messages are end-to-end encrypted via Vault Transit</span>
                  ) : (
                    <span className="encryption-text warning">Messages are not encrypted</span>
                  )}
                </div>
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
                <div className={`encryption-badge ${chatState.encryptionEnabled ? 'encrypted' : 'unencrypted'}`}>
                  <span className="lock-icon">
                    <Icon 
                      name={chatState.encryptionEnabled ? 'lock' : 'unlock'} 
                      size={14} 
                      color={chatState.encryptionEnabled ? 'var(--color-success)' : 'var(--color-warning)'} 
                    />
                  </span>
                  <span className="encryption-text">
                    {chatState.encryptionEnabled ? 'Encrypted' : 'Not Encrypted'}
                  </span>
                </div>
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
                        <span className="time">
                          {new Date(message.timestamp).toLocaleTimeString([], { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })}
                        </span>
                      </div>
                      <div className="message-body">
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
            </div>
          </>
        ) : (
          <div className="no-room">
            <p>Select a room to start messaging</p>
          </div>
        )}
      </main>

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
