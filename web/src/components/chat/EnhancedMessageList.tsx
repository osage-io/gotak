import React, { useEffect, useRef, useState } from 'react';
import type { ChatMessage, ReactionType } from '../../types/chat';
import { locationSharingService, type SharedLocation } from '../../services/locationSharingService';
import { formatCoordinates, getRelativeTime } from '../../utils/common';
import './EnhancedMessageList.css';

interface EnhancedMessageListProps {
  roomId: string;
  messages: ChatMessage[];
  currentUserId: string;
  onReaction: (messageId: string, reactionType: ReactionType) => void;
  onAcknowledge: (messageId: string) => void;
  onReply: (messageId: string) => void;
  onLocationClick?: (location: SharedLocation, message: ChatMessage) => void;
  onShowOnMap?: (location: SharedLocation, message: ChatMessage) => void;
  typingUsers?: Set<string>;
}

export const EnhancedMessageList: React.FC<EnhancedMessageListProps> = ({
  messages,
  currentUserId,
  onReaction,
  onAcknowledge,
  onReply,
  onLocationClick,
  onShowOnMap,
  typingUsers = new Set(),
}) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [expandedLocations, setExpandedLocations] = useState<Set<string>>(new Set());
  
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
  
  const toggleLocationExpanded = (messageId: string) => {
    setExpandedLocations(prev => {
      const newSet = new Set(prev);
      if (newSet.has(messageId)) {
        newSet.delete(messageId);
      } else {
        newSet.add(messageId);
      }
      return newSet;
    });
  };
  
  const handleLocationClick = (message: ChatMessage) => {
    const location = locationSharingService.extractLocationFromMessage(message);
    if (location && onLocationClick) {
      onLocationClick(location, message);
    }
  };
  
  const handleShowOnMap = (message: ChatMessage) => {
    const location = locationSharingService.extractLocationFromMessage(message);
    if (location && onShowOnMap) {
      onShowOnMap(location, message);
    }
  };
  
  const renderLocationInfo = (message: ChatMessage) => {
    const location = locationSharingService.extractLocationFromMessage(message);
    if (!location) return null;
    
    const isExpanded = expandedLocations.has(message.id);
    const displayText = locationSharingService.formatLocationForDisplay(location);
    
    return (
      <div className="message-location-info">
        <div 
          className="location-header"
          onClick={() => toggleLocationExpanded(message.id)}
        >
          <div className="location-display">
            <span className="location-icon">📍</span>
            <span className="location-coords">{displayText}</span>
          </div>
          <button className="location-expand-btn" title="Toggle details">
            {isExpanded ? '▼' : '▶'}
          </button>
        </div>
        
        {isExpanded && (
          <div className="location-details">
            <div className="location-info-grid">
              <div className="location-info-item">
                <span className="info-label">Coordinates:</span>
                <span className="info-value">{formatCoordinates(location.lat, location.lng, 6)}</span>
              </div>
              
              {location.alt && (
                <div className="location-info-item">
                  <span className="info-label">Altitude:</span>
                  <span className="info-value">{Math.round(location.alt)}m</span>
                </div>
              )}
              
              {location.accuracy && (
                <div className="location-info-item">
                  <span className="info-label">Accuracy:</span>
                  <span className="info-value">±{Math.round(location.accuracy)}m</span>
                </div>
              )}
              
              <div className="location-info-item">
                <span className="info-label">Source:</span>
                <span className="info-value capitalize">{location.source.replace('_', ' ')}</span>
              </div>
              
              <div className="location-info-item">
                <span className="info-label">Time:</span>
                <span className="info-value">{getRelativeTime(location.timestamp)}</span>
              </div>
            </div>
            
            <div className="location-actions">
              <button
                className="location-action-btn"
                onClick={() => handleLocationClick(message)}
                title="Go to location"
              >
                🎯 Go to
              </button>
              
              {onShowOnMap && (
                <button
                  className="location-action-btn"
                  onClick={() => handleShowOnMap(message)}
                  title="Show on map"
                >
                  🗺️ Show on map
                </button>
              )}
              
              <button
                className="location-action-btn"
                onClick={() => {
                  const coordText = `${location.lat}, ${location.lng}`;
                  navigator.clipboard?.writeText(coordText);
                }}
                title="Copy coordinates"
              >
                📋 Copy
              </button>
            </div>
          </div>
        )}
      </div>
    );
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
    const hasLocation = Boolean(message.locationLat && message.locationLng);
    
    return (
      <div
        key={message.id}
        className={`enhanced-message ${isOwnMessage ? 'enhanced-message--own' : 'enhanced-message--other'} ${hasLocation ? 'enhanced-message--has-location' : ''}`}
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
            {message.requiresAck && !message.isAcknowledged && (
              <span className="ack-required-indicator" title="Acknowledgment required">✋</span>
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
          
          {hasLocation && renderLocationInfo(message)}
          
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
          
          {hasLocation && (
            <button
              className="action-btn location-btn"
              onClick={() => handleShowOnMap(message)}
              title="Show location on map"
            >
              🗺️
            </button>
          )}
          
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
            <div className="reaction-options">
              {(['roger', 'wilco', 'negative', 'like', 'important', 'question'] as ReactionType[]).map(type => (
                <button
                  key={type}
                  className="reaction-option"
                  onClick={() => onReaction(message.id, type)}
                  title={type}
                >
                  {getReactionIcon(type)}
                </button>
              ))}
            </div>
          </div>
        </div>
        
        {message.acknowledgments && message.acknowledgments.length > 0 && (
          <div className="message-acknowledgments">
            <span className="ack-label">Acknowledged by:</span>
            {message.acknowledgments.map((ack, index) => (
              <span key={ack.id} className="ack-user">
                {ack.callsign || ack.username}
                {index < message.acknowledgments!.length - 1 && ', '}
              </span>
            ))}
          </div>
        )}
      </div>
    );
  };
  
  const renderTypingUsers = () => {
    if (typingUsers.size === 0) return null;
    
    const userList = Array.from(typingUsers).join(', ');
    const text = typingUsers.size === 1 
      ? `${userList} is typing...` 
      : `${userList} are typing...`;
    
    return (
      <div className="typing-indicator">
        <div className="typing-dots">
          <span></span>
          <span></span>
          <span></span>
        </div>
        <span className="typing-text">{text}</span>
      </div>
    );
  };
  
  return (
    <div className="enhanced-message-list">
      <div className="messages-container">
        {messages.map(renderMessage)}
        {renderTypingUsers()}
        <div ref={messagesEndRef} />
      </div>
    </div>
  );
};

export default EnhancedMessageList;
