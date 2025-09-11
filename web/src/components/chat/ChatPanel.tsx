import { useState } from 'react';
import { useChat } from '../../hooks/useChat';
import { RoomsList } from './RoomsList';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import { UsersList } from './UsersList';
import { CreateRoomModal } from './CreateRoomModal';
import type { ChatPanelProps, ReactionType } from '../../types/chat';
import './ChatPanel.css';

export function ChatPanel({ isVisible, onToggle }: ChatPanelProps) {
  const [showCreateRoom, setShowCreateRoom] = useState(false);
  const [showUsersList, setShowUsersList] = useState(false);
  
  // Mock current user ID - in a real app this would come from auth
  const currentUserId = 'current-user-123';
  
  const {
    rooms,
    messages,
    currentRoom,
    typingUsers,
    loading,
    error,
    selectRoom,
    createRoom,
    leaveRoom,
    sendMessage,
    acknowledgeMessage,
    addReaction,
    sendTypingIndicator,
    isConnected,
  } = useChat(currentUserId);

  const currentRoomMessages = currentRoom ? messages[currentRoom.id] || [] : [];
  const currentRoomTyping = currentRoom ? typingUsers[currentRoom.id] || new Set() : new Set();

  const handleSendMessage = (messageText: string, replyToId?: string) => {
    if (!currentRoom) return;

    sendMessage({
      roomId: currentRoom.id,
      messageText,
      messageType: 'text',
      priority: 'normal',
      classification: 'UNCLASSIFIED',
      replyToId,
      requiresAck: false,
    });
  };

  const handleReaction = (messageId: string, reactionType: ReactionType) => {
    addReaction(messageId, reactionType);
  };

  const handleAcknowledge = (messageId: string) => {
    acknowledgeMessage(messageId);
  };

  const handleReply = (messageId: string) => {
    // This will be handled by the MessageInput component
    console.log('Reply to message:', messageId);
  };

  const handleTyping = (typing: boolean) => {
    if (currentRoom) {
      sendTypingIndicator(currentRoom.id, typing);
    }
  };

  if (!isVisible) {
    return (
      <div className="chat-panel chat-panel--collapsed">
        <button 
          className="chat-toggle-btn"
          onClick={onToggle}
          title="Open Chat"
        >
          💬
          {rooms.reduce((total, room) => total + (room.unreadCount || 0), 0) > 0 && (
            <span className="unread-badge">
              {rooms.reduce((total, room) => total + (room.unreadCount || 0), 0)}
            </span>
          )}
        </button>
      </div>
    );
  }

  return (
    <div className="chat-panel chat-panel--expanded">
      <div className="chat-header">
        <h3>Tactical Chat</h3>
        <div className="chat-header-controls">
          <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            <span className="status-indicator"></span>
            <span className="status-text">{isConnected ? 'Connected' : 'Disconnected'}</span>
          </div>
          {currentRoom && (
            <button 
              className="users-btn"
              onClick={() => setShowUsersList(!showUsersList)}
              title="Show participants"
            >
              👥 {currentRoom.participantCount}
            </button>
          )}
          <button 
            className="close-btn"
            onClick={onToggle}
            title="Close Chat"
          >
            ✕
          </button>
        </div>
      </div>

      <div className="chat-body">
        {error && (
          <div className="chat-error">
            <span>⚠️ {error}</span>
          </div>
        )}

        <div className="chat-layout">
          <div className="chat-sidebar">
            <RoomsList
              rooms={rooms}
              selectedRoomId={currentRoom?.id}
              onSelectRoom={selectRoom}
              onCreateRoom={() => setShowCreateRoom(true)}
              onLeaveRoom={leaveRoom}
              loading={loading}
            />
          </div>

          <div className="chat-main">
            {currentRoom ? (
              <>
                <div className="chat-room-header">
                  <div className="room-info">
                    <h4>{currentRoom.name}</h4>
                    <span className="room-classification">{currentRoom.classification}</span>
                    <span className="room-type">{currentRoom.type}</span>
                  </div>
                  {currentRoom.description && (
                    <p className="room-description">{currentRoom.description}</p>
                  )}
                </div>

                <MessageList
                  roomId={currentRoom.id}
                  messages={currentRoomMessages}
                  currentUserId={currentUserId}
                  onReaction={handleReaction}
                  onAcknowledge={handleAcknowledge}
                  onReply={handleReply}
                  typingUsers={new Set(Array.from(currentRoomTyping).map(u => String(u)))}
                />

                <MessageInput
                  roomId={currentRoom.id}
                  onSendMessage={handleSendMessage}
                  onTyping={handleTyping}
                  disabled={!isConnected}
                />
              </>
            ) : (
              <div className="no-room-selected">
                <h4>Select a chat room to start messaging</h4>
                <p>Choose a room from the sidebar or create a new one.</p>
              </div>
            )}
          </div>

          {showUsersList && currentRoom && (
            <div className="chat-users-sidebar">
              <UsersList
                participants={currentRoom.participants || []}
                currentUserId={currentUserId}
                onRemoveUser={(userId) => {
                  // TODO: Implement remove user functionality
                  console.log('Remove user:', userId);
                }}
                onChangeRole={(userId, role) => {
                  // TODO: Implement change role functionality
                  console.log('Change role:', userId, role);
                }}
              />
            </div>
          )}
        </div>
      </div>

      {showCreateRoom && (
        <CreateRoomModal
          onCreateRoom={async (roomData) => {
            await createRoom(roomData);
            setShowCreateRoom(false);
          }}
          onCancel={() => setShowCreateRoom(false)}
        />
      )}
    </div>
  );
}
