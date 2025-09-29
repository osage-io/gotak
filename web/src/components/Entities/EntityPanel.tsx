/**
 * Entity Panel Component (Stub)
 */

import React, { memo } from 'react';

export const EntityPanel: React.FC = memo(() => {
  return (
    <div style={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: '#0a0d10', color: '#dcddde' }}>
      <div style={{ textAlign: 'center', padding: '2rem' }}>
        <h2 style={{ color: '#00d4aa', marginBottom: '1rem' }}>👥 Entity Management</h2>
        <p style={{ color: '#b0b3b8' }}>Entity management panel will be implemented here</p>
      </div>
    </div>
  );
});

EntityPanel.displayName = 'EntityPanel';
