import type { ChatRoom } from '../../types/chat';

interface RoomsListProps {
  rooms: ChatRoom[];
  selectedRoomId?: string;
  onSelectRoom: (roomId: string) => void;
  onCreateRoom: () => void;
  onLeaveRoom: (roomId: string) => void;
  loading?: boolean;
}

export function RoomsList({
  rooms,
  selectedRoomId,
  onSelectRoom,
  onCreateRoom,
  onLeaveRoom,
  loading = false,
}: RoomsListProps) {
  const getRoomTypeIcon = (type: string) => {
    switch (type) {
      case 'tactical':
        return '🎯';
      case 'emergency':
        return '🚨';
      case 'private':
        return '🔒';
      case 'group':
      default:
        return '👥';
    }
  };

  const getClassificationColor = (classification: string) => {
    switch (classification) {
      case 'UNCLASSIFIED':
        return 'classification-unclassified';
      case 'RESTRICTED':
        return 'classification-restricted';
      case 'CONFIDENTIAL':
        return 'classification-confidential';
      case 'SECRET':
        return 'classification-secret';
      case 'TOP_SECRET':
        return 'classification-top-secret';
      default:
        return 'classification-unclassified';
    }
  };

  const formatLastMessageTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 1) {
      return 'now';
    } else if (diffInMinutes < 60) {
      return `${diffInMinutes}m`;
    } else if (diffInMinutes < 1440) {
      return `${Math.floor(diffInMinutes / 60)}h`;
    } else {
      return date.toLocaleDateString();
    }
  };

  return (
    <div className="rooms-list">
      <div className="rooms-header">
        <h4>Chat Rooms</h4>
        <button
          className="create-room-btn"
          onClick={onCreateRoom}
          title="Create new room"
        >
          ➕
        </button>
      </div>

      {loading && (
        <div className="rooms-loading">
          <span>Loading rooms...</span>
        </div>
      )}

      {rooms.length === 0 && !loading && (
        <div className="no-rooms">
          <p>No chat rooms available.</p>
          <button onClick={onCreateRoom} className="create-first-room-btn">
            Create your first room
          </button>
        </div>
      )}

      <div className="rooms-container">
        {rooms.map((room) => (
          <div
            key={room.id}
            className={`room-item ${selectedRoomId === room.id ? 'selected' : ''}`}
            onClick={() => onSelectRoom(room.id)}
          >
            <div className="room-item-header">
              <div className="room-title">
                <span className="room-icon">{getRoomTypeIcon(room.type)}</span>
                <span className="room-name">{room.name}</span>
                {room.unreadCount > 0 && (
                  <span className="unread-count">{room.unreadCount}</span>
                )}
              </div>
              <div className="room-actions">
                <button
                  className="leave-room-btn"
                  onClick={(e) => {
                    e.stopPropagation();
                    if (window.confirm(`Leave room "${room.name}"?`)) {
                      onLeaveRoom(room.id);
                    }
                  }}
                  title="Leave room"
                >
                  ❌
                </button>
              </div>
            </div>

            <div className="room-meta">
              <span className={`room-classification ${getClassificationColor(room.classification)}`}>
                {room.classification}
              </span>
              <span className="room-participants">
                👥 {room.participantCount}
              </span>
              {room.lastMessage && (
                <span className="last-message-time">
                  {formatLastMessageTime(room.lastMessage.createdAt)}
                </span>
              )}
            </div>

            {room.lastMessage && (
              <div className="last-message">
                <span className="last-message-sender">
                  {room.lastMessage.senderCallsign || room.lastMessage.senderUsername || 'Unknown'}:
                </span>
                <span className="last-message-text">
                  {room.lastMessage.messageText.length > 50
                    ? `${room.lastMessage.messageText.substring(0, 50)}...`
                    : room.lastMessage.messageText}
                </span>
              </div>
            )}

            {room.description && (
              <div className="room-description">
                {room.description}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
