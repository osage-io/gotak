import React, { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Icon } from './ui/Icon';
import { Message } from '../types/comms';
import aiService from '../services/aiService';
import './AIChat.css';

interface AIChatProps {
  onClose?: () => void;
}

const AIChat: React.FC<AIChatProps> = ({ onClose }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Vault-aware: the key may live only in Vault KV, which isConfigured() can't see.
  const [configured, setConfigured] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Quick action buttons
  const quickActions = [
    { iconName: 'target', label: 'Mission Brief', action: 'Provide current mission briefing and objectives' },
    { iconName: 'map-pin', label: 'Area Intel', action: 'Give me area intelligence for my current position' },
    { iconName: 'shield', label: 'Threat Assessment', action: 'What are the current threat levels in the AO?' },
    { iconName: 'cloud', label: 'Weather', action: 'Provide tactical weather report for operations' },
  ];

  useEffect(() => {
    // Initial greeting from AI Intel Officer
    const initialMessage: Message = {
      id: Date.now().toString(),
      content: 'Intel Officer online. Ready to provide tactical support. How can I assist your operation?',
      sender: 'AI Intel Officer',
      timestamp: new Date().toISOString(),
      type: 'text',
      room: 'ai-intel'
    };
    setMessages([initialMessage]);
  }, []);

  // Check configuration against Vault (key may not be in localStorage).
  useEffect(() => {
    let active = true;
    aiService.ensureConfigured().then(ok => { if (active) setConfigured(ok); });
    return () => { active = false; };
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async () => {
    if (!inputValue.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      content: inputValue,
      sender: 'Field Operator',
      timestamp: new Date().toISOString(),
      type: 'text',
      room: 'ai-intel'
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);
    setError(null);

    try {
      const response = await aiService.sendMessage(inputValue, messages);
      
      const aiMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: response.content,
        sender: 'AI Intel Officer',
        timestamp: new Date().toISOString(),
        type: 'text',
        room: 'ai-intel'
      };

      setMessages(prev => [...prev, aiMessage]);
    } catch (err) {
      console.error('Failed to get AI response:', err);
      const msg = err instanceof Error ? err.message : String(err);
      // Surface the real cause so the operator knows what to fix.
      const authFailed = /401|authentication|invalid x-api-key|invalid.*key/i.test(msg);
      const banner = authFailed
        ? 'API key rejected by Anthropic. Update it in Integrations → Anthropic.'
        : 'Connection to Intel Officer failed. Check secure comms channel.';
      const chatMsg = authFailed
        ? 'AUTH FAILURE: Anthropic rejected the API key (401). Reconfigure a valid key in Integrations → Anthropic.'
        : `COMMS ERROR: ${msg || 'Unable to establish secure connection.'} Retry transmission.`;
      setError(banner);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: chatMsg,
        sender: 'System',
        timestamp: new Date().toISOString(),
        type: 'system',
        room: 'ai-intel'
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickAction = (action: string) => {
    setInputValue(action);
    inputRef.current?.focus();
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { 
      hour12: false, 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <div className="ai-chat-container">
      <div className="ai-chat-header">
        <div className="ai-chat-header-info">
          <div className="ai-chat-header-icon">
            <Icon name="bot" size={32} color="#00ff41" />
          </div>
          <div>
            <h3>AI Intelligence Officer</h3>
            <span className="ai-chat-status">
              <span className="status-indicator active"></span>
              Secure Channel Active
            </span>
          </div>
        </div>
        {onClose && (
          <button className="ai-chat-close" onClick={onClose}>×</button>
        )}
      </div>

      {!configured && (
        <div className="ai-chat-config-prompt">
          <div className="config-prompt-icon">
            <Icon name="settings" size={24} color="#ff6b35" />
          </div>
          <h4>AI Intel Officer Setup Required</h4>
          <p>Configure Anthropic API to enable AI Intelligence Officer capabilities</p>
          <button 
            className="config-prompt-button"
            onClick={() => window.location.href = '/integrations?setup=anthropic'}
          >
            <Icon name="wrench" size={16} color="#000000" />
            <span>Configure Integration</span>
          </button>
        </div>
      )}

      <div className="ai-chat-quick-actions">
        {quickActions.map((action, index) => (
          <button
            key={index}
            className="quick-action-btn"
            onClick={() => handleQuickAction(action.action)}
            disabled={isLoading}
          >
            <Icon name={action.iconName} size={16} color="#00ff41" />
            <span>{action.label}</span>
          </button>
        ))}
      </div>

      <div className="ai-chat-messages">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`ai-chat-message ${
              message.sender === 'AI Intel Officer' ? 'ai' : 
              message.sender === 'System' ? 'system' : 'user'
            }`}
          >
            <div className="message-header">
              <span className="message-sender">{message.sender}</span>
              <span className="message-time">{formatTime(message.timestamp)}</span>
            </div>
            <div className="message-content">
              {message.sender === 'AI Intel Officer' ? (
                <div className="markdown-body">
                  <ReactMarkdown>{message.content}</ReactMarkdown>
                </div>
              ) : (
                message.content
              )}
            </div>
          </div>
        ))}
        {isLoading && (
          <div className="ai-chat-message ai loading">
            <div className="message-header">
              <span className="message-sender">AI Intel Officer</span>
            </div>
            <div className="message-content">
              <span className="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="ai-chat-input-container">
        <input
          ref={inputRef}
          type="text"
          className="ai-chat-input"
          placeholder="Enter message for Intel Officer..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyPress={handleKeyPress}
          disabled={isLoading || !configured}
        />
        <button
          className="ai-chat-send"
          onClick={handleSendMessage}
          disabled={!inputValue.trim() || isLoading || !configured}
        >
          <Icon name="send" size={20} color="#000000" />
        </button>
      </div>

      {error && (
        <div className="ai-chat-error">
          <Icon name="alert-circle" size={16} color="#ff4444" />
          <span>{error}</span>
        </div>
      )}
    </div>
  );
};

export default AIChat;