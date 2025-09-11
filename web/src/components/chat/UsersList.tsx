import type { ChatRoomParticipant, ParticipantRole } from '../../types/chat';

interface UsersListProps {
  participants: ChatRoomParticipant[];
  currentUserId: string;
  onRemoveUser?: (userId: string) => void;
  onChangeRole?: (userId: string, role: ParticipantRole) => void;
}

export function UsersList({
  participants,
  currentUserId,
  onRemoveUser,
  onChangeRole,
}: UsersListProps) {
  const getRoleIcon = (role: ParticipantRole) => {
    switch (role) {
      case 'admin':
        return '👑';
      case 'moderator':
        return '🛡️';
      case 'member':
        return '👤';
      case 'observer':
        return '👁️';
      default:
        return '👤';
    }
  };

  const getRoleName = (role: ParticipantRole) => {
    switch (role) {
      case 'admin':
        return 'Admin';
      case 'moderator':
        return 'Moderator';
      case 'member':
        return 'Member';
      case 'observer':
        return 'Observer';
      default:
        return 'Member';
    }
  };

  const getRoleColor = (role: ParticipantRole) => {
    switch (role) {
      case 'admin':
        return 'role-admin';
      case 'moderator':
        return 'role-moderator';
      case 'member':
        return 'role-member';
      case 'observer':
        return 'role-observer';
      default:
        return 'role-member';
    }
  };

  const formatJoinDate = (joinedAt: string) => {
    const date = new Date(joinedAt);
    const now = new Date();
    const diffInDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
    
    if (diffInDays === 0) {
      return 'Today';
    } else if (diffInDays === 1) {
      return 'Yesterday';
    } else if (diffInDays < 7) {
      return `${diffInDays} days ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  const formatLastSeen = (lastSeen: string) => {
    const date = new Date(lastSeen);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 5) {
      return 'Online';
    } else if (diffInMinutes < 60) {
      return `${diffInMinutes}m ago`;
    } else if (diffInMinutes < 1440) {
      return `${Math.floor(diffInMinutes / 60)}h ago`;
    } else {
      return `${Math.floor(diffInMinutes / 1440)}d ago`;
    }
  };

  const currentUserParticipant = participants.find(p => p.userId === currentUserId);
  const canManageUsers = currentUserParticipant?.role === 'admin' || currentUserParticipant?.role === 'moderator';

  // Sort participants by role hierarchy, then by name
  const sortedParticipants = [...participants].sort((a, b) => {
    const roleOrder = { admin: 0, moderator: 1, member: 2, observer: 3 };
    const roleComparison = roleOrder[a.role] - roleOrder[b.role];
    
    if (roleComparison !== 0) {
      return roleComparison;
    }
    
    const nameA = a.callsign || a.username || a.displayName || 'Unknown';
    const nameB = b.callsign || b.username || b.displayName || 'Unknown';
    return nameA.localeCompare(nameB);
  });

  return (
    <div className="users-list">
      <div className="users-header">
        <h4>Participants ({participants.length})</h4>
      </div>

      <div className="users-container">
        {sortedParticipants.map((participant) => {
          const isCurrentUser = participant.userId === currentUserId;
          const displayName = participant.callsign || participant.username || participant.displayName || 'Unknown';
          const isOnline = formatLastSeen(participant.lastSeen) === 'Online';

          return (
            <div key={participant.id} className={`user-item ${isCurrentUser ? 'current-user' : ''}`}>
              <div className="user-avatar">
                <div className={`avatar-circle ${isOnline ? 'online' : 'offline'}`}>
                  <span className="avatar-text">
                    {displayName.charAt(0).toUpperCase()}
                  </span>
                  <div className={`status-indicator ${isOnline ? 'online' : 'offline'}`}></div>
                </div>
              </div>

              <div className="user-info">
                <div className="user-header">
                  <span className="user-name">
                    {displayName}
                    {isCurrentUser && <span className="you-indicator"> (You)</span>}
                  </span>
                  <div className="user-actions">
                    {canManageUsers && !isCurrentUser && onChangeRole && (
                      <select
                        className="role-select"
                        value={participant.role}
                        onChange={(e) => onChangeRole(participant.userId!, e.target.value as ParticipantRole)}
                        title="Change role"
                      >
                        <option value="observer">Observer</option>
                        <option value="member">Member</option>
                        <option value="moderator">Moderator</option>
                        {currentUserParticipant?.role === 'admin' && (
                          <option value="admin">Admin</option>
                        )}
                      </select>
                    )}
                    {canManageUsers && !isCurrentUser && onRemoveUser && (
                      <button
                        className="remove-user-btn"
                        onClick={() => {
                          if (window.confirm(`Remove ${displayName} from this room?`)) {
                            onRemoveUser(participant.userId!);
                          }
                        }}
                        title="Remove user"
                      >
                        🚫
                      </button>
                    )}
                  </div>
                </div>

                <div className="user-meta">
                  <div className={`user-role ${getRoleColor(participant.role)}`}>
                    <span className="role-icon">{getRoleIcon(participant.role)}</span>
                    <span className="role-name">{getRoleName(participant.role)}</span>
                  </div>
                  <div className="user-status">
                    <span className="last-seen">{formatLastSeen(participant.lastSeen)}</span>
                  </div>
                </div>

                <div className="user-details">
                  {participant.username && participant.username !== displayName && (
                    <div className="user-username">@{participant.username}</div>
                  )}
                  {participant.displayName && participant.displayName !== displayName && (
                    <div className="user-display-name">{participant.displayName}</div>
                  )}
                  <div className="joined-date">
                    Joined {formatJoinDate(participant.joinedAt)}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {participants.length === 0 && (
        <div className="no-users">
          <p>No participants in this room.</p>
        </div>
      )}
    </div>
  );
}
