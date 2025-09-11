import { useState, useRef, useEffect } from 'react';

interface MessageInputProps {
  roomId: string;
  onSendMessage: (messageText: string, replyToId?: string) => void;
  onTyping: (typing: boolean) => void;
  disabled?: boolean;
  replyTo?: { id: string; senderName: string; messageText: string };
  onCancelReply?: () => void;
}

export function MessageInput({
  onSendMessage,
  onTyping,
  disabled = false,
  replyTo,
  onCancelReply,
}: MessageInputProps) {
  const [message, setMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
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

    // Clear previous timeout and set new one
    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current);
    }

    // Stop typing indicator after 2 seconds of inactivity
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

    onSendMessage(trimmedMessage, replyTo?.id);
    setMessage('');
    
    // Stop typing indicator
    if (isTyping) {
      setIsTyping(false);
      onTyping(false);
    }

    // Clear timeout
    if (typingTimeoutRef.current) {
      window.clearTimeout(typingTimeoutRef.current);
    }

    // Clear reply
    if (onCancelReply) {
      onCancelReply();
    }

    // Focus back to input
    if (textareaRef.current) {
      textareaRef.current.focus();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (e.shiftKey) {
        // Allow new line with Shift+Enter
        return;
      } else {
        // Send message with Enter
        e.preventDefault();
        handleSubmit(e);
      }
    } else if (e.key === 'Escape' && replyTo && onCancelReply) {
      // Cancel reply with Escape
      onCancelReply();
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
    <div className="message-input">
      {replyTo && (
        <div className="reply-context">
          <div className="reply-info">
            <span className="reply-indicator">↳ Replying to</span>
            <strong className="reply-sender">{replyTo.senderName}:</strong>
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
                className="option-btn"
                title="Add location"
                disabled={disabled}
              >
                📍
              </button>
              <button
                type="button"
                className="option-btn"
                title="Set priority"
                disabled={disabled}
              >
                ⚠️
              </button>
              <button
                type="button"
                className="option-btn"
                title="Require acknowledgment"
                disabled={disabled}
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
          </div>
        </div>
      </form>
    </div>
  );
}
