import { useEffect, useRef } from 'react';
import type { ChatMessage, ReactionType } from '../../types/chat';

interface MessageListProps {
  roomId: string;
  messages: ChatMessage[];
  currentUserId: string;
  onReaction: (messageId: string, reactionType: ReactionType) => void;
  onAcknowledge: (messageId: string) => void;
  onReply: (messageId: string) => void;
  typingUsers?: Set<string>;
}

export function MessageList({
  messages,
  currentUserId,
  onReaction,
  onAcknowledge,
  onReply,
  typingUsers = new Set(),
}: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const formatMessageTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const isToday = date.toDateString() === now.toDateString();
    
    if (isToday) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' }) + 
             ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
  };

  const getPriorityIcon = (priority: string) => {
    switch (priority) {
      case 'emergency':
        return '🚨';
      case 'urgent':
        return '❗';
      case 'high':
        return '⚠️';
      case 'low':
        return '🔽';
      default:
        return null;
    }
  };

  const getMessageTypeIcon = (type: string) => {
    switch (type) {
      case 'system':
        return '🤖';
      case 'position':
        return '📍';
      case 'emergency':
        return '🚨';
      case 'tactical_report':
        return '📋';
      default:
        return null;
    }
  };

  const getReactionIcon = (reactionType: ReactionType) => {
    switch (reactionType) {
      case 'roger':
        return '✅';
      case 'wilco':
        return '🫡';
      case 'negative':
        return '❌';
      case 'like':
        return '👍';
      case 'important':
        return '⭐';
      case 'question':
        return '❓';
      default:
        return '👍';
    }
  };

  const renderReactions = (message: ChatMessage) => {
    if (!message.reactions || message.reactions.length === 0) {
      return null;
    }

    // Group reactions by type
    const reactionGroups = message.reactions.reduce((groups, reaction) => {
      if (!groups[reaction.reactionType]) {
        groups[reaction.reactionType] = [];
      }
      groups[reaction.reactionType].push(reaction);
      return groups;
    }, {} as Record<ReactionType, typeof message.reactions>);

    return (
      <div className="message-reactions">
        {Object.entries(reactionGroups).map(([type, reactions]) => (
          <button
            key={type}
            className={`reaction-btn ${reactions.some(r => r.userId === currentUserId) ? 'reaction-btn--active' : ''}`}
            onClick={() => onReaction(message.id, type as ReactionType)}
            title={reactions.map(r => r.username || r.callsign || 'Unknown').join(', ')}
          >
            {getReactionIcon(type as ReactionType)} {reactions.length}
          </button>
        ))}
      </div>
    );
  };

  const renderMessage = (message: ChatMessage) => {
    const isOwnMessage = message.senderId === currentUserId;
    const senderName = message.senderCallsign || message.senderUsername || 'Unknown';
    const priorityIcon = getPriorityIcon(message.priority);
    const typeIcon = getMessageTypeIcon(message.messageType);

    return (
      <div
        key={message.id}
        className={`message ${isOwnMessage ? 'message--own' : 'message--other'}`}
      >
        <div className="message-header">
          <span className="message-sender">{senderName}</span>
          <span className="message-time">{formatMessageTime(message.createdAt)}</span>
          <div className="message-indicators">
            {priorityIcon && <span className="priority-indicator" title={message.priority}>{priorityIcon}</span>}
            {typeIcon && <span className="type-indicator" title={message.messageType}>{typeIcon}</span>}
            {message.classification !== 'UNCLASSIFIED' && (
              <span className={`classification-indicator classification-${message.classification.toLowerCase()}`}>
                {message.classification}
              </span>
            )}
          </div>
        </div>

        {message.replyTo && (
          <div className="message-reply-context">
            <span className="reply-indicator">↳</span>
            <span className="reply-to">
              {message.replyTo.senderCallsign || message.replyTo.senderUsername}: {message.replyTo.messageText}
            </span>
          </div>
        )}

        <div className="message-content">
          <p className="message-text">{message.messageText}</p>
          
          {message.locationLat && message.locationLng && (
            <div className="message-location">
              📍 Location: {message.locationLat.toFixed(6)}, {message.locationLng.toFixed(6)}
            </div>
          )}

          {message.keywords && message.keywords.length > 0 && (
            <div className="message-keywords">
              {message.keywords.map((keyword, index) => (
                <span key={index} className="keyword-tag">#{keyword}</span>
              ))}
            </div>
          )}
        </div>

        {renderReactions(message)}

        <div className="message-actions">
          <button
            className="action-btn"
            onClick={() => onReaction(message.id, 'roger')}
            title="Acknowledge"
          >
            ✅
          </button>
          <button
            className="action-btn"
            onClick={() => onReply(message.id)}
            title="Reply"
          >
            💬
          </button>
          {message.requiresAck && !message.isAcknowledged && (
            <button
              className="action-btn acknowledge-btn"
              onClick={() => onAcknowledge(message.id)}
              title="Acknowledge message"
            >
              ✋ ACK
            </button>
          )}
          <div className="reaction-picker">
            <button className="action-btn" title="Add reaction">😀</button>
            <div className="reaction-menu">
              {(['roger', 'wilco', 'negative', 'like', 'important', 'question'] as ReactionType[]).map((reactionType) => (
                <button
                  key={reactionType}
                  className="reaction-option"
                  onClick={() => onReaction(message.id, reactionType)}
                  title={reactionType}
                >
                  {getReactionIcon(reactionType)}
                </button>
              ))}
            </div>
          </div>
        </div>

        {message.acknowledgments && message.acknowledgments.length > 0 && (
          <div className="message-acknowledgments">
            <span className="ack-label">Acknowledged by:</span>
            {message.acknowledgments.map((ack, index) => (
              <span key={index} className="ack-user">
                {ack.username || ack.callsign || 'Unknown'}
              </span>
            ))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="message-list">
      <div className="messages-container">
        {messages.length === 0 ? (
          <div className="no-messages">
            <p>No messages yet. Start the conversation!</p>
          </div>
        ) : (
          messages.map(renderMessage)
        )}

        {typingUsers.size > 0 && (
          <div className="typing-indicators">
            <div className="typing-message">
              <span className="typing-users">
                {Array.from(typingUsers).join(', ')} {typingUsers.size === 1 ? 'is' : 'are'} typing...
              </span>
              <div className="typing-dots">
                <span></span>
                <span></span>
                <span></span>
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>
    </div>
  );
}
