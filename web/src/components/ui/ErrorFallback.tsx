/**
 * Error Fallback Component
 * Used with ErrorBoundary to display error states
 */

import React from 'react';
import './ErrorFallback.css';

interface ErrorFallbackProps {
  error: Error;
  resetErrorBoundary: () => void;
}

const ErrorFallback: React.FC<ErrorFallbackProps> = ({ error, resetErrorBoundary }) => {
  return (
    <div className="error-fallback">
      <div className="error-icon">⚠️</div>
      <h2>Something went wrong</h2>
      <details className="error-details">
        <summary>Error Details</summary>
        <pre>{error.message}</pre>
        {error.stack && (
          <pre className="error-stack">{error.stack}</pre>
        )}
      </details>
      <button className="retry-button" onClick={resetErrorBoundary}>
        Try Again
      </button>
    </div>
  );
};

export default ErrorFallback;
