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
<<<<<<< HEAD
        const response = await fetch('/api/health', {
=======
        // We'll use the /health endpoint at the root level, not under /api
        const response = await fetch('/health', {
>>>>>>> temp-merge-fix
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          // Short timeout to avoid long waiting times
          signal: AbortSignal.timeout(3000),
        });
<<<<<<< HEAD
        
=======

>>>>>>> temp-merge-fix
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

<<<<<<< HEAD
    // Set up periodic checks
    const intervalId = setInterval(checkApiAvailability, 30000); // Check every 30 seconds
=======
    // Set up periodic checks - only if there's an error
    // This reduces unnecessary requests to the backend
    const intervalId = setInterval(() => {
      if (apiError) {
        checkApiAvailability();
      }
    }, 60000); // Check every 60 seconds, but only if there's an error
>>>>>>> temp-merge-fix

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
<<<<<<< HEAD
                  If you're a developer, make sure the backend server is running at http://localhost:8080
=======
                  If you're a developer, make sure the Go backend server is running at http://localhost:8080
                  <br />
                  Run <code className="bg-gray-100 px-1 rounded">cd backend && go run main.go</code> to start the server.
>>>>>>> temp-merge-fix
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
