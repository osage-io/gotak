import React, { useState, useRef, useEffect } from 'react';
import type { SendMessageRequest, MessagePriority, ChatMessage } from '../../types/chat';
import { locationSharingService, type SharedLocation, type LocationShareOptions } from '../../services/locationSharingService';
import './EnhancedMessageInput.css';

interface EnhancedMessageInputProps {
  roomId: string;
  onSendMessage: (message: SendMessageRequest) => void;
  onTyping: (typing: boolean) => void;
  disabled?: boolean;
  replyTo?: ChatMessage;
  onCancelReply?: () => void;
  currentLocation?: SharedLocation;
}

interface LocationInputState {
  isOpen: boolean;
  useCurrentLocation: boolean;
  customLocation: { lat: string; lng: string; alt: string };
  customMessage: string;
  priority: MessagePriority;
  requiresAck: boolean;
}

export const EnhancedMessageInput: React.FC<EnhancedMessageInputProps> = ({
  roomId,
  onSendMessage,
  onTyping,
  disabled = false,
  replyTo,
  onCancelReply,
  currentLocation,
}) => {
  const [message, setMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [priority, setPriority] = useState<MessagePriority>('normal');
  const [requiresAck, setRequiresAck] = useState(false);
  
  // Location sharing state
  const [locationInput, setLocationInput] = useState<LocationInputState>({
    isOpen: false,
    useCurrentLocation: true,
    customLocation: { lat: '', lng: '', alt: '' },
    customMessage: '',
    priority: 'normal',
    requiresAck: false,
  });
  
  // Loading states
  const [gettingLocation, setGettingLocation] = useState(false);
  const [locationError, setLocationError] = useState<string | null>(null);
  
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const typingTimeoutRef = useRef<number | null>(null);
  
  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px';
    }
  }, [message]);
  
  // Focus textarea when reply is set
  useEffect(() => {
    if (replyTo && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [replyTo]);
  
  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setMessage(value);
    
    // Handle typing indicators
    if (value.length > 0 && !isTyping) {
      setIsTyping(true);
      onTyping(true);
    }
    
    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current);
    }
    
    typingTimeoutRef.current = window.setTimeout(() => {
      if (isTyping) {
        setIsTyping(false);
        onTyping(false);
      }
    }, 2000);
  };
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const trimmedMessage = message.trim();
    if (!trimmedMessage || disabled) return;
    
    const messageRequest: SendMessageRequest = {
      roomId,
      messageText: trimmedMessage,
      messageType: 'text',
      priority,
      classification: 'UNCLASSIFIED',
      replyToId: replyTo?.id,
      requiresAck,
    };
    
    onSendMessage(messageRequest);
    setMessage('');
    setPriority('normal');
    setRequiresAck(false);
    
    // Stop typing indicator
    if (isTyping) {
      setIsTyping(false);
      onTyping(false);
    }
    
    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current);
    }
    
    if (onCancelReply) {
      onCancelReply();
    }
    
    if (textareaRef.current) {
      textareaRef.current.focus();
    }
  };
  
  const handleLocationShare = async () => {
    if (locationInput.useCurrentLocation) {
      await shareCurrentLocation();
    } else {
      await shareCustomLocation();
    }
  };
  
  const shareCurrentLocation = async () => {
    setGettingLocation(true);
    setLocationError(null);
    
    try {
      const location = currentLocation || await locationSharingService.getCurrentLocation();
      
      const options: LocationShareOptions = {
        includeAltitude: true,
        requiresAck: locationInput.requiresAck,
        priority: locationInput.priority,
        customMessage: locationInput.customMessage || undefined,
      };
      
      const locationMessage = locationSharingService.createLocationMessage(roomId, location, options);
      onSendMessage(locationMessage);
      
      setLocationInput(prev => ({ ...prev, isOpen: false, customMessage: '' }));
    } catch (error) {
      setLocationError(error instanceof Error ? error.message : 'Failed to get location');
    } finally {
      setGettingLocation(false);
    }
  };
  
  const shareCustomLocation = async () => {
    const { lat, lng, alt } = locationInput.customLocation;
    
    if (!lat || !lng) {
      setLocationError('Please enter valid coordinates');
      return;
    }
    
    const latitude = parseFloat(lat);
    const longitude = parseFloat(lng);
    const altitude = alt ? parseFloat(alt) : undefined;
    
    if (!locationSharingService.validateLocation(latitude, longitude)) {
      setLocationError('Invalid coordinates. Latitude must be -90 to 90, Longitude must be -180 to 180');
      return;
    }
    
    const location = locationSharingService.createLocationFromInput(latitude, longitude, altitude);
    
    const options: LocationShareOptions = {
      includeAltitude: Boolean(altitude),
      requiresAck: locationInput.requiresAck,
      priority: locationInput.priority,
      customMessage: locationInput.customMessage || undefined,
    };
    
    const locationMessage = locationSharingService.createLocationMessage(roomId, location, options);
    onSendMessage(locationMessage);
    
    setLocationInput(prev => ({ 
      ...prev, 
      isOpen: false, 
      customMessage: '',
      customLocation: { lat: '', lng: '', alt: '' }
    }));
    setLocationError(null);
  };
  
  const handleParseLocationFromText = () => {
    const location = locationSharingService.parseLocationFromText(message);
    if (location) {
      setLocationInput(prev => ({
        ...prev,
        isOpen: true,
        useCurrentLocation: false,
        customLocation: {
          lat: location.lat.toString(),
          lng: location.lng.toString(),
          alt: location.alt?.toString() || '',
        },
      }));
    }
  };
  
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (e.shiftKey) {
        return;
      } else {
        e.preventDefault();
        handleSubmit(e);
      }
    } else if (e.key === 'Escape') {
      if (locationInput.isOpen) {
        setLocationInput(prev => ({ ...prev, isOpen: false }));
      } else if (replyTo && onCancelReply) {
        onCancelReply();
      }
    }
  };
  
  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        window.clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);
  
  return (
    <div className="enhanced-message-input">
      {replyTo && (
        <div className="reply-context">
          <div className="reply-info">
            <span className="reply-indicator">↳ Replying to</span>
            <strong className="reply-sender">
              {replyTo.senderCallsign || replyTo.senderUsername || 'Unknown'}:
            </strong>
            <span className="reply-text">
              {replyTo.messageText.length > 100 
                ? `${replyTo.messageText.substring(0, 100)}...` 
                : replyTo.messageText}
            </span>
          </div>
          {onCancelReply && (
            <button 
              className="cancel-reply-btn"
              onClick={onCancelReply}
              title="Cancel reply"
            >
              ✕
            </button>
          )}
        </div>
      )}
      
      {/* Location sharing panel */}
      {locationInput.isOpen && (
        <div className="location-input-panel">
          <div className="location-panel-header">
            <h4>Share Location</h4>
            <button 
              className="close-location-btn"
              onClick={() => setLocationInput(prev => ({ ...prev, isOpen: false }))}
            >
              ✕
            </button>
          </div>
          
          <div className="location-options">
            <label className="location-option">
              <input
                type="radio"
                checked={locationInput.useCurrentLocation}
                onChange={(e) => setLocationInput(prev => ({ 
                  ...prev, 
                  useCurrentLocation: e.target.checked 
                }))}
              />
              <span className="option-label">📍 Use current GPS location</span>
            </label>
            
            <label className="location-option">
              <input
                type="radio"
                checked={!locationInput.useCurrentLocation}
                onChange={(e) => setLocationInput(prev => ({ 
                  ...prev, 
                  useCurrentLocation: !e.target.checked 
                }))}
              />
              <span className="option-label">🗺️ Enter coordinates manually</span>
            </label>
          </div>
          
          {!locationInput.useCurrentLocation && (
            <div className="custom-location-inputs">
              <div className="coordinate-inputs">
                <input
                  type="number"
                  step="any"
                  placeholder="Latitude"
                  value={locationInput.customLocation.lat}
                  onChange={(e) => setLocationInput(prev => ({
                    ...prev,
                    customLocation: { ...prev.customLocation, lat: e.target.value }
                  }))}
                  className="coordinate-input"
                />
                <input
                  type="number"
                  step="any"
                  placeholder="Longitude"
                  value={locationInput.customLocation.lng}
                  onChange={(e) => setLocationInput(prev => ({
                    ...prev,
                    customLocation: { ...prev.customLocation, lng: e.target.value }
                  }))}
                  className="coordinate-input"
                />
                <input
                  type="number"
                  step="any"
                  placeholder="Altitude (optional)"
                  value={locationInput.customLocation.alt}
                  onChange={(e) => setLocationInput(prev => ({
                    ...prev,
                    customLocation: { ...prev.customLocation, alt: e.target.value }
                  }))}
                  className="coordinate-input"
                />
              </div>
              
              <button
                type="button"
                className="parse-location-btn"
                onClick={handleParseLocationFromText}
                title="Parse location from message text"
              >
                📝 Parse from text
              </button>
            </div>
          )}
          
          <div className="location-message-options">
            <textarea
              placeholder="Optional custom message..."
              value={locationInput.customMessage}
              onChange={(e) => setLocationInput(prev => ({ 
                ...prev, 
                customMessage: e.target.value 
              }))}
              className="location-message-input"
              rows={2}
            />
            
            <div className="location-options-row">
              <select
                value={locationInput.priority}
                onChange={(e) => setLocationInput(prev => ({ 
                  ...prev, 
                  priority: e.target.value as MessagePriority 
                }))}
                className="priority-select"
              >
                <option value="low">Low Priority</option>
                <option value="normal">Normal</option>
                <option value="high">High Priority</option>
                <option value="urgent">Urgent</option>
              </select>
              
              <label className="ack-checkbox">
                <input
                  type="checkbox"
                  checked={locationInput.requiresAck}
                  onChange={(e) => setLocationInput(prev => ({ 
                    ...prev, 
                    requiresAck: e.target.checked 
                  }))}
                />
                Require ACK
              </label>
            </div>
          </div>
          
          {locationError && (
            <div className="location-error">
              ⚠️ {locationError}
            </div>
          )}
          
          <div className="location-actions">
            <button
              type="button"
              className="location-action-btn secondary"
              onClick={() => setLocationInput(prev => ({ ...prev, isOpen: false }))}
            >
              Cancel
            </button>
            <button
              type="button"
              className="location-action-btn primary"
              onClick={handleLocationShare}
              disabled={gettingLocation}
            >
              {gettingLocation ? '📡 Getting location...' : '📍 Share Location'}
            </button>
          </div>
        </div>
      )}
      
      <form onSubmit={handleSubmit} className="message-form">
        <div className="input-container">
          <textarea
            ref={textareaRef}
            value={message}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder={disabled ? "Disconnected..." : "Type a message... (Enter to send, Shift+Enter for new line)"}
            disabled={disabled}
            className="message-textarea"
            rows={1}
            maxLength={4000}
          />
          
          <div className="input-actions">
            <div className="message-options">
              <button
                type="button"
                className={`option-btn ${locationInput.isOpen ? 'active' : ''}`}
                title="Share location"
                disabled={disabled}
                onClick={() => setLocationInput(prev => ({ ...prev, isOpen: !prev.isOpen }))}
              >
                📍
              </button>
              
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as MessagePriority)}
                className="priority-select-small"
                title="Message priority"
                disabled={disabled}
              >
                <option value="low">🔽</option>
                <option value="normal">➖</option>
                <option value="high">⚠️</option>
                <option value="urgent">❗</option>
              </select>
              
              <button
                type="button"
                className={`option-btn ${requiresAck ? 'active' : ''}`}
                title="Require acknowledgment"
                disabled={disabled}
                onClick={() => setRequiresAck(!requiresAck)}
              >
                ✋
              </button>
            </div>
            
            <button
              type="submit"
              className="send-btn"
              disabled={disabled || !message.trim()}
              title="Send message"
            >
              📤
            </button>
          </div>
        </div>
        
        <div className="input-footer">
          <div className="character-count">
            {message.length}/4000
          </div>
          <div className="input-hints">
            <span>Enter to send • Shift+Enter for new line</span>
            {priority !== 'normal' && (
              <span className="priority-indicator">
                Priority: {priority.toUpperCase()}
              </span>
            )}
            {requiresAck && (
              <span className="ack-indicator">ACK Required</span>
            )}
          </div>
        </div>
      </form>
    </div>
  );
};

export default EnhancedMessageInput;
