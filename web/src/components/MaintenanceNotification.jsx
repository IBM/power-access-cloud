import React, { useState, useEffect, useCallback } from "react";
import { InlineNotification } from "@carbon/react";
import { getAllMaintenanceWindows } from "../services/request";
import "../styles/maintenance-notification.scss";

// Constants
const CHECK_INTERVAL_MS = 600000; // 10 minutes
const CACHE_DURATION_MS = 600000; // 10 minutes cache
const DEFAULT_MESSAGE = "Scheduled maintenance in progress. Some services may be temporarily unavailable.";

// Cache for API response
let cachedWindows = null;
let cacheTimestamp = null;

const MaintenanceNotification = () => {
  const [activeWindow, setActiveWindow] = useState(null);
  const [loading, setLoading] = useState(true);

  // Fetch active maintenance window from API
  // Backend now returns only the earliest active window
  const fetchMaintenanceWindow = useCallback(async () => {
    try {
      // Check cache first
      const now = Date.now();
      if (cachedWindows && cacheTimestamp && (now - cacheTimestamp) < CACHE_DURATION_MS) {
        return cachedWindows;
      }

      // Fetch from API (backend returns only active window)
      const response = await getAllMaintenanceWindows();
      
      if (response.type === "GET_ALL_MAINTENANCE_WINDOWS" && response.payload) {
        cachedWindows = response.payload.maintenances || [];
        cacheTimestamp = now;
        return cachedWindows;
      }
      
      // Return empty array if API fails
      return [];
    } catch (error) {
      return [];
    }
  }, []);

  // Check for active maintenance window
  const checkMaintenanceWindow = useCallback(async () => {
    setLoading(true);
    
    try {
      const windows = await fetchMaintenanceWindow();
      
      // Backend returns only 1 active window (or empty array)
      if (windows && windows.length > 0) {
        const window = windows[0];
        setActiveWindow({
          id: window.id,
          message: window.message || DEFAULT_MESSAGE,
        });
      } else {
        setActiveWindow(null);
      }
      
      setLoading(false);
    } catch (error) {
      setActiveWindow(null);
      setLoading(false);
    }
  }, [fetchMaintenanceWindow]);

  useEffect(() => {
    // Check immediately on mount
    checkMaintenanceWindow();

    // Set up interval to check periodically
    const interval = setInterval(checkMaintenanceWindow, CHECK_INTERVAL_MS);

    // Cleanup interval on unmount
    return () => clearInterval(interval);
  }, [checkMaintenanceWindow]);

  // Handle closing notification
  const handleClose = () => {
    setActiveWindow(null);
  };

  // Don't render if no active window or still loading
  if (loading || !activeWindow) {
    return null;
  }

  return (
    <div className="maintenance-notification-container">
      <InlineNotification
        kind="warning"
        title="Maintenance Notice"
        subtitle={activeWindow.message}
        hideCloseButton={false}
        onCloseButtonClick={handleClose}
        lowContrast
        aria-live="polite"
        role="alert"
      />
    </div>
  );
};

export default MaintenanceNotification;
