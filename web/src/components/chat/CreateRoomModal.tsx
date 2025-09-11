import React, { useState } from 'react';
import type { CreateChatRoomRequest, ChatRoomType, Classification } from '../../types/chat';

interface CreateRoomModalProps {
  onCreateRoom: (room: CreateChatRoomRequest) => Promise<void>;
  onCancel: () => void;
}

export function CreateRoomModal({ onCreateRoom, onCancel }: CreateRoomModalProps) {
  const [formData, setFormData] = useState<CreateChatRoomRequest>({
    name: '',
    description: '',
    type: 'group',
    classification: 'UNCLASSIFIED',
    settings: {},
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const roomTypes: { value: ChatRoomType; label: string; description: string; icon: string }[] = [
    { value: 'group', label: 'Group Chat', description: 'General purpose group discussion', icon: '👥' },
    { value: 'tactical', label: 'Tactical', description: 'Mission-critical communications', icon: '🎯' },
    { value: 'emergency', label: 'Emergency', description: 'Emergency response coordination', icon: '🚨' },
    { value: 'private', label: 'Private', description: 'Direct message channel', icon: '🔒' },
  ];

  const classifications: { value: Classification; label: string; color: string }[] = [
    { value: 'UNCLASSIFIED', label: 'Unclassified', color: '#22c55e' },
    { value: 'RESTRICTED', label: 'Restricted', color: '#3b82f6' },
    { value: 'CONFIDENTIAL', label: 'Confidential', color: '#f59e0b' },
    { value: 'SECRET', label: 'Secret', color: '#ef4444' },
    { value: 'TOP_SECRET', label: 'Top Secret', color: '#7c3aed' },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!formData.name.trim()) {
      setError('Room name is required');
      return;
    }

    if (formData.name.length < 2) {
      setError('Room name must be at least 2 characters');
      return;
    }

    if (formData.name.length > 255) {
      setError('Room name must be less than 255 characters');
      return;
    }

    if (formData.description && formData.description.length > 1000) {
      setError('Description must be less than 1000 characters');
      return;
    }

    try {
      setIsSubmitting(true);
      
      // Prepare room data
      const roomData: CreateChatRoomRequest = {
        ...formData,
        name: formData.name.trim(),
        description: formData.description?.trim() || undefined,
        settings: {
          maxParticipants: 100,
          allowFileUploads: false,
          requireAcknowledgment: formData.type === 'emergency' || formData.type === 'tactical',
          ...formData.settings,
        },
      };

      await onCreateRoom(roomData);
    } catch (err) {
      setError(`Failed to create room: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleInputChange = (field: keyof CreateChatRoomRequest, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  const selectedRoomType = roomTypes.find(rt => rt.value === formData.type);
  const selectedClassification = classifications.find(c => c.value === formData.classification);

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-content create-room-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Create New Chat Room</h3>
          <button className="modal-close-btn" onClick={onCancel} title="Close">
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="create-room-form">
          {error && (
            <div className="form-error">
              <span>⚠️ {error}</span>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="room-name" className="form-label">
              Room Name *
            </label>
            <input
              id="room-name"
              type="text"
              className="form-input"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              placeholder="Enter room name"
              maxLength={255}
              required
              autoFocus
            />
            <div className="form-hint">
              {formData.name.length}/255 characters
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="room-description" className="form-label">
              Description
            </label>
            <textarea
              id="room-description"
              className="form-textarea"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              placeholder="Optional room description"
              rows={3}
              maxLength={1000}
            />
            <div className="form-hint">
              {(formData.description || '').length}/1000 characters
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Room Type *</label>
            <div className="room-type-grid">
              {roomTypes.map((type) => (
                <label
                  key={type.value}
                  className={`room-type-option ${formData.type === type.value ? 'selected' : ''}`}
                >
                  <input
                    type="radio"
                    name="room-type"
                    value={type.value}
                    checked={formData.type === type.value}
                    onChange={(e) => handleInputChange('type', e.target.value as ChatRoomType)}
                    className="room-type-radio"
                  />
                  <div className="room-type-content">
                    <div className="room-type-header">
                      <span className="room-type-icon">{type.icon}</span>
                      <span className="room-type-label">{type.label}</span>
                    </div>
                    <div className="room-type-description">{type.description}</div>
                  </div>
                </label>
              ))}
            </div>
            {selectedRoomType && (
              <div className="selected-type-info">
                <strong>{selectedRoomType.icon} {selectedRoomType.label}:</strong> {selectedRoomType.description}
              </div>
            )}
          </div>

          <div className="form-group">
            <label className="form-label">Security Classification *</label>
            <div className="classification-grid">
              {classifications.map((classification) => (
                <label
                  key={classification.value}
                  className={`classification-option ${formData.classification === classification.value ? 'selected' : ''}`}
                >
                  <input
                    type="radio"
                    name="classification"
                    value={classification.value}
                    checked={formData.classification === classification.value}
                    onChange={(e) => handleInputChange('classification', e.target.value as Classification)}
                    className="classification-radio"
                  />
                  <div className="classification-content">
                    <div
                      className="classification-indicator"
                      style={{ backgroundColor: classification.color }}
                    ></div>
                    <span className="classification-label">{classification.label}</span>
                  </div>
                </label>
              ))}
            </div>
            {selectedClassification && (
              <div className="selected-classification-info">
                <strong>Classification Level:</strong> {selectedClassification.label}
                <br />
                <small>Ensure all participants have appropriate security clearance.</small>
              </div>
            )}
          </div>

          <div className="form-actions">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={onCancel}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isSubmitting || !formData.name.trim()}
            >
              {isSubmitting ? (
                <>
                  <span className="spinner">⏳</span>
                  Creating...
                </>
              ) : (
                'Create Room'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
