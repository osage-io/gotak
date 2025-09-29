/**
 * Loading Spinner Component
 * Reusable spinner with different sizes and optional message
 */

import React from 'react';
import './LoadingSpinner.css';

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  message?: string;
  className?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  size = 'medium', 
  message,
  className = '' 
}) => {
  return (
    <div className={`loading-spinner ${size} ${className}`}>
      <div className="spinner-ring">
        <div></div>
        <div></div>
        <div></div>
        <div></div>
      </div>
      {message && <p className="loading-message">{message}</p>}
    </div>
  );
};

export default LoadingSpinner;
