/**
 * Communications Page - TAK Chat & Messaging
 * Real-time tactical communication system with CoT message integration
 */

import React, { useState, useEffect, useRef } from 'react';
import { wsService } from '../services/websocketService';
import { ChatMessage, ChatRoom } from '../services/apiClient';

interface ChatState {
  rooms: ChatRoom[];
  activeRoomId: string | null;
  messages: Map<string, ChatMessage[]>;
  typingUsers: Map<string, Set<string>>;
}

const Communications: React.FC = () => {
  const [chatState, setChatState] = useState<ChatState>({
    rooms: [],
    activeRoomId: null,
    messages: new Map(),
    typingUsers: new Map(),
  });

  const [messageInput, setMessageInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [showNewRoomDialog, setShowNewRoomDialog] = useState(false);
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();

  // Auto-scroll to bottom of messages
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [chatState.messages, chatState.activeRoomId]);

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
        name: 'Tactical Operations',
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
    ];

    setChatState(prev => ({
      ...prev,
      rooms: initialRooms,
      activeRoomId: 'general',
      messages: new Map(initialRooms.map(room => [room.id, []])),
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
          roomTypingUsers.add(`${username} (${userId})`);
        } else {
          roomTypingUsers.delete(`${username} (${userId})`);
        }
        
        newTypingUsers.set(roomId, roomTypingUsers);
        return { ...prev, typingUsers: newTypingUsers };
      });
    });

    const unsubscribeRoomUpdate = wsService.onChatRoomUpdate(({ room, action }) => {
      setChatState(prev => {
        let newRooms = [...prev.rooms];
        
        switch (action) {
          case 'created':
            newRooms.push(room);
            break;
          case 'updated':
            const index = newRooms.findIndex(r => r.id === room.id);
            if (index >= 0) {
              newRooms[index] = room;
            }
            break;
        }
        
        return { ...prev, rooms: newRooms };
      });
    });

    return () => {
      unsubscribeChatMessage();
      unsubscribeUserTyping();
      unsubscribeRoomUpdate();
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

    const success = wsService.sendChatMessage(chatState.activeRoomId, messageInput.trim(), {
      messageType: 'text',
      priority: 'normal',
      classification: 'UNCLASSIFIED',
    });

    if (success) {
      setMessageInput('');
      setIsTyping(false);
      if (chatState.activeRoomId) {
        wsService.setTyping(chatState.activeRoomId, false);
      }
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

  // Join room
  const joinRoom = (roomId: string) => {
    wsService.joinChatRoom(roomId);
    setChatState(prev => ({ ...prev, activeRoomId: roomId }));
  };

  // Create new room
  const createNewRoom = () => {
    if (!newRoomName.trim()) return;

    // Mock room creation - in real app, call API
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
    wsService.joinChatRoom(newRoom.id);
  };

  // Get message priority color
  const getPriorityColor = (priority: ChatMessage['priority']) => {
    switch (priority) {
      case 'urgent': return 'var(--color-error)';
      case 'high': return 'var(--color-warning)';
      case 'normal': return 'var(--color-text-primary)';
      case 'low': return 'var(--color-text-muted)';
      default: return 'var(--color-text-primary)';
    }
  };

  // Get classification badge color
  const getClassificationColor = (classification: ChatMessage['classification']) => {
    switch (classification) {
      case 'TOP_SECRET': return 'var(--color-error)';
      case 'SECRET': return 'var(--color-warning)';
      case 'CONFIDENTIAL': return 'var(--color-info)';
      case 'UNCLASSIFIED': return 'var(--color-success)';
      default: return 'var(--color-neutral)';
    }
  };

  const activeRoom = chatState.rooms.find(r => r.id === chatState.activeRoomId);
  const activeMessages = chatState.activeRoomId ? 
    chatState.messages.get(chatState.activeRoomId) || [] : [];
  const typingInActiveRoom = chatState.activeRoomId ?
    chatState.typingUsers.get(chatState.activeRoomId) || new Set() : new Set();

  return (
    <div className="page-container">
      {/* Sidebar - Chat Rooms */}
      <aside className="chat-sidebar">
        <div className="sidebar-header">
          <h2 className="font-display font-semibold text-lg text-primary">
            Communications
          </h2>
          <button
            onClick={() => setShowNewRoomDialog(true)}
            className="btn-ghost text-xs"
            style={{ padding: 'var(--space-2) var(--space-3)' }}
          >
            + New Room
          </button>
        </div>

        <div className="rooms-list">
          {chatState.rooms.map((room) => (
            <button
              key={room.id}
              onClick={() => joinRoom(room.id)}
              className={`room-item ${room.id === chatState.activeRoomId ? 'active' : ''}`}
            >
              <div className="room-info">
                <div className="room-name font-medium">
                  {room.name}
                </div>
                <div className="room-type text-xs text-muted uppercase">
                  {room.type}
                </div>
              </div>
              <div className="room-indicator">
                {room.type === 'private' ? '🔒' : '🌐'}
              </div>
            </button>
          ))}
        </div>
      </aside>

      {/* Main Chat Area */}
      <main className="chat-main">
        {activeRoom ? (
          <>
            {/* Chat Header */}
            <header className="chat-header">
              <div className="room-title">
                <h3 className="font-display font-semibold text-xl text-primary">
                  #{activeRoom.name}
                </h3>
                <div className="room-meta font-mono text-sm text-secondary">
                  {activeRoom.type === 'private' ? '🔒 Private Room' : '🌐 Public Room'}
                  {/* • {activeRoom.members.length} members */}
                </div>
              </div>
              
              <div className="chat-actions">
                <button className="btn-ghost text-xs">Settings</button>
                <button className="btn-ghost text-xs">Members</button>
              </div>
            </header>

            {/* Messages Area */}
            <div className="messages-container">
              <div className="messages-list">
                {activeMessages.map((message) => (
                  <div key={message.id} className="message-item">
                    <div className="message-header">
                      <div className="message-sender font-semibold text-sm">
                        {message.senderCallsign || message.senderId}
                      </div>
                      <div className="message-meta">
                        {message.classification && (
                          <span 
                            className="classification-badge"
                            style={{ backgroundColor: getClassificationColor(message.classification) }}
                          >
                            {message.classification}
                          </span>
                        )}
                        <span className="message-time font-mono text-xs text-muted">
                          {new Date(message.createdAt).toLocaleTimeString()}
                        </span>
                      </div>
                    </div>
                    <div 
                      className="message-content text-sm"
                      style={{ color: getPriorityColor(message.priority) }}
                    >
                      {message.messageText}
                    </div>
                    {message.locationLat && message.locationLng && (
                      <div className="message-location font-mono text-xs text-info">
                        📍 {message.locationLat.toFixed(6)}, {message.locationLng.toFixed(6)}
                      </div>
                    )}
                  </div>
                ))}

                {/* Typing Indicators */}
                {typingInActiveRoom.size > 0 && (
                  <div className="typing-indicator">
                    <div className="typing-dots">
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                    <div className="typing-text text-xs text-muted">
                      {Array.from(typingInActiveRoom).join(', ')} 
                      {typingInActiveRoom.size === 1 ? ' is' : ' are'} typing...
                    </div>
                  </div>
                )}
                
                <div ref={messagesEndRef} />
              </div>
            </div>

            {/* Message Input */}
            <div className="message-input-container">
              <div className="input-wrapper">
                <textarea
                  value={messageInput}
                  onChange={handleInputChange}
                  onKeyPress={handleKeyPress}
                  placeholder={`Message #${activeRoom.name}...`}
                  className="message-input"
                  rows={3}
                />
                <div className="input-actions">
                  <div className="input-meta text-xs text-muted">
                    Press Enter to send, Shift+Enter for new line
                  </div>
                  <button
                    onClick={sendMessage}
                    disabled={!messageInput.trim()}
                    className="btn-primary"
                  >
                    Send
                  </button>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="no-room-selected">
            <div className="text-muted text-center">
              <h3 className="font-display text-xl">Select a room to start chatting</h3>
              <p className="text-sm">Choose from the available rooms in the sidebar</p>
            </div>
          </div>
        )}
      </main>

      {/* New Room Dialog */}
      {showNewRoomDialog && (
        <div className="modal-overlay" onClick={() => setShowNewRoomDialog(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="font-display font-semibold text-lg text-primary">
                Create New Room
              </h3>
              <button
                onClick={() => setShowNewRoomDialog(false)}
                className="btn-ghost"
                style={{ padding: 'var(--space-1)' }}
              >
                ✕
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label className="form-label">Room Name</label>
                <input
                  type="text"
                  value={newRoomName}
                  onChange={(e) => setNewRoomName(e.target.value)}
                  placeholder="Enter room name..."
                  className="form-input"
                  autoFocus
                />
              </div>
            </div>
            <div className="modal-actions">
              <button
                onClick={() => setShowNewRoomDialog(false)}
                className="btn-ghost"
              >
                Cancel
              </button>
              <button
                onClick={createNewRoom}
                disabled={!newRoomName.trim()}
                className="btn-primary"
              >
                Create Room
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Communications Styles */}
      <style jsx>{`
        .page-container {
          display: flex;
          flex-direction: row;
          gap: var(--spacing-lg);
          height: 100%;
          padding: 0;
        }

        .chat-sidebar {
          width: 320px;
          background: linear-gradient(180deg, 
            rgba(26, 31, 38, 0.95) 0%, 
            rgba(15, 20, 25, 0.95) 100%);
          border: 1px solid rgba(0, 212, 170, 0.15);
          border-radius: var(--radius-lg);
          display: flex;
          flex-direction: column;
          box-shadow: 
            0 4px 24px rgba(0, 0, 0, 0.4),
            0 0 0 1px rgba(0, 212, 170, 0.08),
            inset 0 1px 0 rgba(255, 255, 255, 0.03);
          backdrop-filter: blur(10px);
          overflow: hidden;
        }

        .sidebar-header {
          padding: var(--spacing-xl);
          border-bottom: 1px solid rgba(0, 212, 170, 0.15);
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.08) 0%, 
            rgba(0, 0, 0, 0.3) 100%);
          display: flex;
          justify-content: space-between;
          align-items: center;
          backdrop-filter: blur(8px);
        }

        .rooms-list {
          flex: 1;
          padding: var(--space-4);
          display: flex;
          flex-direction: column;
          gap: var(--space-2);
          overflow-y: auto;
        }

        .room-item {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: var(--spacing-md) var(--spacing-lg);
          background: linear-gradient(135deg, 
            transparent 0%, 
            rgba(0, 0, 0, 0.1) 100%);
          border: 1px solid rgba(255, 255, 255, 0.05);
          border-radius: var(--radius-lg);
          color: var(--color-text-secondary);
          cursor: pointer;
          transition: all 0.25s cubic-bezier(0.25, 0.8, 0.25, 1);
          text-align: left;
          position: relative;
          overflow: hidden;
        }
        
        .room-item::before {
          content: '';
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          width: 3px;
          background: linear-gradient(180deg, 
            var(--color-accent) 0%, 
            rgba(0, 212, 170, 0.6) 100%);
          transform: scaleY(0);
          transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
          transform-origin: bottom;
        }

        .room-item:hover {
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.08) 0%, 
            rgba(0, 0, 0, 0.2) 100%);
          border-color: rgba(0, 212, 170, 0.2);
          color: var(--color-text-primary);
          transform: translateX(4px);
          box-shadow: 
            0 2px 8px rgba(0, 0, 0, 0.2),
            0 0 0 1px rgba(0, 212, 170, 0.05);
        }
        
        .room-item:hover::before {
          transform: scaleY(1);
        }

        .room-item.active {
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.15) 0%, 
            rgba(0, 0, 0, 0.3) 100%);
          border-color: rgba(0, 212, 170, 0.3);
          color: var(--color-accent);
          box-shadow: 
            0 4px 12px rgba(0, 0, 0, 0.3),
            0 0 0 1px rgba(0, 212, 170, 0.1),
            inset 0 1px 0 rgba(255, 255, 255, 0.05);
        }
        
        .room-item.active::before {
          transform: scaleY(1);
        }

        .room-info {
          flex: 1;
        }

        .room-name {
          margin-bottom: var(--space-1);
        }

        .room-indicator {
          opacity: 0.6;
        }

        .chat-main {
          flex: 1;
          display: flex;
          flex-direction: column;
          overflow: hidden;
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.6) 0%, 
            rgba(15, 20, 25, 0.8) 100%);
          border: 1px solid rgba(0, 212, 170, 0.15);
          border-radius: var(--radius-lg);
          box-shadow: 
            0 4px 24px rgba(0, 0, 0, 0.4),
            0 0 0 1px rgba(0, 212, 170, 0.08);
          backdrop-filter: blur(8px);
        }

        .chat-header {
          padding: var(--spacing-xl) var(--spacing-xxl);
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.06) 0%, 
            rgba(0, 0, 0, 0.3) 100%);
          border-bottom: 1px solid rgba(0, 212, 170, 0.15);
          display: flex;
          justify-content: space-between;
          align-items: center;
          backdrop-filter: blur(8px);
        }

        .room-meta {
          margin-top: var(--space-1);
        }

        .chat-actions {
          display: flex;
          gap: var(--space-2);
        }

        .messages-container {
          flex: 1;
          overflow: hidden;
          position: relative;
        }

        .messages-list {
          height: 100%;
          overflow-y: auto;
          padding: var(--space-4) var(--space-8);
          display: flex;
          flex-direction: column;
          gap: var(--space-4);
        }

        .message-item {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.7) 0%, 
            rgba(15, 20, 25, 0.5) 100%);
          border: 1px solid rgba(0, 212, 170, 0.1);
          border-left: 3px solid rgba(0, 212, 170, 0.3);
          border-radius: var(--radius-lg);
          padding: var(--spacing-lg);
          box-shadow: 
            0 2px 8px rgba(0, 0, 0, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.03);
          backdrop-filter: blur(4px);
          transition: all 0.2s ease;
        }
        
        .message-item:hover {
          border-left-color: var(--color-accent);
          transform: translateX(2px);
          box-shadow: 
            0 4px 12px rgba(0, 0, 0, 0.3),
            0 0 0 1px rgba(0, 212, 170, 0.05);
        }

        .message-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: var(--space-2);
        }

        .message-meta {
          display: flex;
          align-items: center;
          gap: var(--space-2);
        }

        .classification-badge {
          padding: var(--space-1) var(--space-2);
          border-radius: var(--radius-sm);
          font-size: var(--font-size-xs);
          font-weight: var(--font-weight-bold);
          color: white;
          text-transform: uppercase;
        }

        .message-content {
          line-height: var(--leading-relaxed);
          white-space: pre-wrap;
        }

        .message-location {
          margin-top: var(--space-2);
          padding-top: var(--space-2);
          border-top: 1px solid var(--color-border);
        }

        .typing-indicator {
          display: flex;
          align-items: center;
          gap: var(--space-2);
          padding: var(--space-2) var(--space-4);
          background-color: var(--color-surface);
          border-radius: var(--radius-lg);
          margin-top: var(--space-2);
        }

        .typing-dots {
          display: flex;
          gap: var(--space-1);
        }

        .typing-dots span {
          width: 4px;
          height: 4px;
          border-radius: 50%;
          background-color: var(--color-text-muted);
          animation: typing 1.5s infinite;
        }

        .typing-dots span:nth-child(2) {
          animation-delay: 0.2s;
        }

        .typing-dots span:nth-child(3) {
          animation-delay: 0.4s;
        }

        @keyframes typing {
          0%, 60%, 100% { opacity: 0.3; }
          30% { opacity: 1; }
        }

        .message-input-container {
          background: linear-gradient(135deg, 
            rgba(0, 212, 170, 0.04) 0%, 
            rgba(0, 0, 0, 0.3) 100%);
          border-top: 1px solid rgba(0, 212, 170, 0.15);
          padding: var(--spacing-lg) var(--spacing-xxl);
          backdrop-filter: blur(8px);
        }

        .input-wrapper {
          background: linear-gradient(135deg, 
            rgba(26, 31, 38, 0.8) 0%, 
            rgba(15, 20, 25, 0.6) 100%);
          border: 1px solid rgba(0, 212, 170, 0.2);
          border-radius: var(--radius-lg);
          overflow: hidden;
          box-shadow: 
            0 2px 8px rgba(0, 0, 0, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.03);
          backdrop-filter: blur(4px);
        }

        .message-input {
          width: 100%;
          padding: var(--space-4);
          background: transparent;
          border: none;
          color: var(--color-text-primary);
          font-family: var(--font-primary);
          font-size: var(--font-size-sm);
          line-height: var(--leading-normal);
          resize: none;
          outline: none;
        }

        .message-input::placeholder {
          color: var(--color-text-muted);
        }

        .input-actions {
          padding: var(--space-3) var(--space-4);
          background-color: var(--color-bg-tertiary);
          display: flex;
          justify-content: space-between;
          align-items: center;
          border-top: 1px solid var(--color-border);
        }

        .no-room-selected {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: var(--space-16);
        }

        .modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background-color: rgba(0, 0, 0, 0.7);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: var(--z-modal);
        }

        .modal-content {
          background-color: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-lg);
          width: 90%;
          max-width: 400px;
          box-shadow: var(--shadow-lg);
        }

        .modal-header {
          padding: var(--space-6);
          border-bottom: 1px solid var(--color-border);
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .modal-body {
          padding: var(--space-6);
        }

        .form-group {
          margin-bottom: var(--space-4);
        }

        .form-label {
          display: block;
          margin-bottom: var(--space-2);
          font-size: var(--font-size-sm);
          font-weight: var(--font-weight-medium);
          color: var(--color-text-primary);
        }

        .form-input {
          width: 100%;
          padding: var(--space-3) var(--space-4);
          background-color: var(--color-bg-primary);
          border: 1px solid var(--color-border);
          border-radius: var(--radius-md);
          color: var(--color-text-primary);
          font-family: var(--font-primary);
          font-size: var(--font-size-sm);
          transition: border-color var(--transition-fast);
        }

        .form-input:focus {
          outline: none;
          border-color: var(--color-border-accent);
        }

        .modal-actions {
          padding: var(--space-6);
          border-top: 1px solid var(--color-border);
          display: flex;
          justify-content: flex-end;
          gap: var(--space-3);
        }

        /* Responsive Design */
        @media (max-width: 768px) {
          .chat-sidebar {
            width: 100%;
            position: absolute;
            z-index: var(--z-overlay);
            transform: translateX(-100%);
            transition: transform var(--transition-base);
          }

          .chat-sidebar.open {
            transform: translateX(0);
          }

          .chat-header {
            padding: var(--space-4) var(--space-6);
          }

          .messages-list {
            padding: var(--space-4);
          }

          .message-input-container {
            padding: var(--space-4);
          }
        }
      `}</style>
    </div>
  );
};

export default Communications;
