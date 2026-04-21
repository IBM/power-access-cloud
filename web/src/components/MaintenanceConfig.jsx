import React, { useState, useEffect } from "react";
import {
  Toggle,
  TextArea,
  Button,
  InlineNotification,
  Loading,
  Form,
  Stack,
} from "@carbon/react";
import { getMaintenanceStatus, updateMaintenanceConfig } from "../services/request";
import "../styles/maintenance-config.scss";

const MaintenanceConfig = () => {
  const [enabled, setEnabled] = useState(false);
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [notification, setNotification] = useState(null);

  // Helper function to format UTC date for datetime-local input
  const formatDateForInput = (isoString) => {
    if (!isoString) return "";
    // datetime-local expects format: YYYY-MM-DDTHH:mm
    // We keep it in UTC by using the ISO string directly
    return isoString.slice(0, 16);
  };

  // Helper function to convert datetime-local value to UTC ISO string
  const convertToUTC = (dateTimeLocal) => {
    if (!dateTimeLocal) return null;
    // Treat the input as UTC time
    return new Date(dateTimeLocal + ':00Z').toISOString();
  };

  // Fetch current configuration
  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const response = await getMaintenanceStatus();
      if (response.type === "GET_MAINTENANCE_STATUS" && response.payload) {
        const config = response.payload;
        setEnabled(config.enabled || false);
        setStartDate(config.start_time ? formatDateForInput(config.start_time) : "");
        setEndDate(config.end_time ? formatDateForInput(config.end_time) : "");
        setMessage(config.message || "");
      }
    } catch (error) {
      console.error("Error fetching maintenance config:", error);
      showNotification("error", "Failed to load maintenance configuration");
    } finally {
      setLoading(false);
    }
  };

  const showNotification = (kind, subtitle) => {
    setNotification({ kind, subtitle });
    setTimeout(() => setNotification(null), 5000);
  };

  const handleSave = async () => {
    // Validation
    if (enabled) {
      if (!startDate || !endDate) {
        showNotification("error", "Start time and end time are required when maintenance is enabled");
        return;
      }
      if (new Date(convertToUTC(endDate)) <= new Date(convertToUTC(startDate))) {
        showNotification("error", "End time must be after start time");
        return;
      }
      if (!message.trim()) {
        showNotification("error", "Message is required when maintenance is enabled");
        return;
      }
    }

    setSaving(true);
    try {
      const payload = {
        enabled,
        start_time: enabled ? convertToUTC(startDate) : null,
        end_time: enabled ? convertToUTC(endDate) : null,
        message: enabled ? message.trim() : "",
      };

      console.log("Sending payload:", payload);

      const response = await updateMaintenanceConfig(payload);
      
      if (response.type === "UPDATE_MAINTENANCE_CONFIG") {
        const successMsg = enabled
          ? "Successfully updated maintenance notification"
          : "Successfully disabled maintenance notification";
        showNotification("success", successMsg);
        // Refresh the config to show updated values
        await fetchConfig();
      } else if (response.type === "API_ERROR") {
        const errorMsg = response.payload?.response?.data?.error || response.payload?.message || "Failed to update maintenance configuration";
        showNotification("error", errorMsg);
      } else {
        showNotification("error", "Failed to update maintenance configuration");
      }
    } catch (error) {
      console.error("Error updating maintenance config:", error);
      const errorMsg = error.response?.data?.error || error.message || "Failed to update maintenance configuration";
      showNotification("error", errorMsg);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="maintenance-config-loading">
        <Loading description="Loading maintenance configuration..." withOverlay={false} />
      </div>
    );
  }

  return (
    <div className="maintenance-config-container">
      <h2>Maintenance Notification Configuration</h2>
      <p className="maintenance-config-description">
        Configure maintenance notifications that will be displayed to all users. 
        Notifications will appear 24 hours before the maintenance start time and remain visible until the end time.
        <strong> All times are in UTC timezone.</strong>
      </p>

      {notification && (
        <InlineNotification
          kind={notification.kind}
          title={notification.kind === "success" ? "Success" : "Error"}
          subtitle={notification.subtitle}
          onCloseButtonClick={() => setNotification(null)}
          lowContrast
        />
      )}

      <Form className="maintenance-config-form">
        <Stack gap={6}>
          <Toggle
            id="maintenance-enabled"
            labelText="Enable Maintenance Notification"
            labelA="Disabled"
            labelB="Enabled"
            toggled={enabled}
            onToggle={(checked) => setEnabled(checked)}
          />

          {enabled && (
            <>
              <div className="date-picker-group">
                <label htmlFor="start-date" className="cds--label">
                  Maintenance Start Time (UTC)
                </label>
                <input
                  id="start-date"
                  type="datetime-local"
                  className="cds--text-input datetime-input"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  onBlur={(e) => setStartDate(e.target.value)}
                  step="60"
                  min={new Date().toISOString().slice(0, 16)}
                />
                <div className="cds--form__helper-text">
                  Format: YYYY-MM-DDTHH:mm (e.g., {new Date().getFullYear()}-04-15T10:00) - UTC timezone
                </div>
              </div>

              <div className="date-picker-group">
                <label htmlFor="end-date" className="cds--label">
                  Maintenance End Time (UTC)
                </label>
                <input
                  id="end-date"
                  type="datetime-local"
                  className="cds--text-input datetime-input"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  onBlur={(e) => setEndDate(e.target.value)}
                  step="60"
                  min={new Date().toISOString().slice(0, 16)}
                />
                <div className="cds--form__helper-text">
                  Format: YYYY-MM-DDTHH:mm (e.g., {new Date().getFullYear()}-04-15T14:00) - UTC timezone
                </div>
              </div>

              <TextArea
                id="maintenance-message"
                labelText="Notification Message"
                placeholder="Enter the maintenance notification message..."
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                rows={4}
                helperText="This message will be displayed to all users during the maintenance window"
              />
            </>
          )}

          <div className="button-group">
            <Button
              kind="primary"
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "Saving..." : "Save Configuration"}
            </Button>
            {enabled && (
              <Button
                kind="secondary"
                onClick={fetchConfig}
                disabled={saving}
              >
                Reset
              </Button>
            )}
          </div>
        </Stack>
      </Form>
    </div>
  );
};

export default MaintenanceConfig;

