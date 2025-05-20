'use client';

import { useState, useEffect } from 'react';
import AlertMessage from './AlertMessage';

interface ApiErrorBoundaryProps {
  children: React.ReactNode;
}

const ApiErrorBoundary = ({ children }: ApiErrorBoundaryProps) => {
  const [apiError, setApiError] = useState<string | null>(null);

  useEffect(() => {
    // Function to check if the backend API is available
    const checkApiAvailability = async () => {
      try {
        // Use a simple fetch to check if the API is available
        const response = await fetch('/api/health', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          // Short timeout to avoid long waiting times
          signal: AbortSignal.timeout(3000),
        });
        
        if (!response.ok) {
          setApiError('Backend API is not responding properly. Some features may not work.');
        } else {
          setApiError(null);
        }
      } catch (error) {
        console.error('API availability check failed:', error);
        setApiError('Cannot connect to the backend server. Please ensure it is running.');
      }
    };

    // Check API availability when component mounts
    checkApiAvailability();

    // Set up periodic checks
    const intervalId = setInterval(checkApiAvailability, 30000); // Check every 30 seconds

    return () => {
      clearInterval(intervalId);
    };
  }, []);

  return (
    <>
      {apiError && (
        <div className="fixed top-0 left-0 right-0 z-50 p-4">
          <AlertMessage
            type="error"
            message={
              <div>
                <p>{apiError}</p>
                <p className="text-xs mt-1">
                  If you're a developer, make sure the backend server is running at http://localhost:8080
                </p>
              </div>
            }
            onClose={() => setApiError(null)}
          />
        </div>
      )}
      {children}
    </>
  );
};

export default ApiErrorBoundary;
