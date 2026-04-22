import React, { useState, useEffect } from "react";
import {
  Button,
  DataTable,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  Modal,
  TextInput,
  TextArea,
  Toggle,
  InlineNotification,
  TableToolbar,
  TableToolbarContent,
  TableToolbarSearch,
  OverflowMenu,
  OverflowMenuItem,
  Pagination,
} from "@carbon/react";
import { Add } from "@carbon/icons-react";
import {
  getAllMaintenanceWindows,
  createMaintenanceWindow,
  updateMaintenanceWindow,
  deleteMaintenanceWindow,
} from "../services/request";
import "../styles/maintenance-manager.scss";

const MaintenanceManager = () => {
  const [windows, setWindows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingWindow, setEditingWindow] = useState(null);
  const [notification, setNotification] = useState(null);
  const [modalError, setModalError] = useState(null);
  const [searchValue, setSearchValue] = useState("");
  
  // Pagination state
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalItems, setTotalItems] = useState(0);

  // Form state
  const [formData, setFormData] = useState({
    enabled: true,
    start_time: "",
    end_time: "",
    message: "",
  });

  // Fetch all maintenance windows (admin view with ?all=true and pagination)
  const fetchWindows = async (currentPage = page, currentPageSize = pageSize) => {
    setLoading(true);
    try {
      // Pass true to get all windows with pagination
      const response = await getAllMaintenanceWindows(true, currentPage, currentPageSize);
      if (response.type === "GET_ALL_MAINTENANCE_WINDOWS" && response.payload) {
        const maintenances = response.payload.maintenances || [];
        setWindows(maintenances);
        setTotalItems(response.payload.total_items || 0);
      }
    } catch (error) {
      console.error("Error fetching maintenance windows:", error);
      showNotification("error", "Failed to load maintenance windows");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWindows();
  }, [page, pageSize]); // eslint-disable-line react-hooks/exhaustive-deps

  // Handle pagination change
  const handlePaginationChange = ({ page: newPage, pageSize: newPageSize }) => {
    setPage(newPage);
    setPageSize(newPageSize);
  };

  // Show notification
  const showNotification = (kind, message) => {
    setNotification({ kind, message });
    setTimeout(() => setNotification(null), 5000);
  };

  // Convert local datetime to UTC ISO string
  const convertToUTC = (localDatetime) => {
    if (!localDatetime) return null;
    const date = new Date(localDatetime);
    return date.toISOString();
  };

  // Convert UTC ISO string to local datetime-local format
  const convertToLocal = (utcString) => {
    if (!utcString) return "";
    const date = new Date(utcString);
    const offset = date.getTimezoneOffset() * 60000;
    const localDate = new Date(date.getTime() - offset);
    return localDate.toISOString().slice(0, 16);
  };

  // Format date for display in local timezone
  const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return date.toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: true
    });
  };

  // Open modal for creating new window
  const handleCreate = () => {
    setEditingWindow(null);
    setModalError(null);
    setFormData({
      enabled: true,
      start_time: "",
      end_time: "",
      message: "",
    });
    setIsModalOpen(true);
  };

  // Open modal for editing existing window
  const handleEdit = (window) => {
    setEditingWindow(window);
    setModalError(null);
    setFormData({
      enabled: window.enabled,
      start_time: convertToLocal(window.start_time),
      end_time: convertToLocal(window.end_time),
      message: window.message,
    });
    setIsModalOpen(true);
  };

  // Handle toggle enable/disable
  const handleToggleEnabled = async (window) => {
    try {
      const payload = {
        enabled: !window.enabled,
        start_time: window.start_time,
        end_time: window.end_time,
        message: window.message,
      };

      const response = await updateMaintenanceWindow(window.id, payload);
      if (response.type === "UPDATE_MAINTENANCE_WINDOW") {
        showNotification(
          "success",
          `Maintenance window ${payload.enabled ? "enabled" : "disabled"} successfully`
        );
        fetchWindows();
      } else if (response.type === "API_ERROR") {
        const errorMsg = response.payload?.response?.data?.error || "Operation failed";
        showNotification("error", errorMsg);
      }
    } catch (error) {
      console.error("Error toggling maintenance window:", error);
      showNotification("error", "Failed to toggle maintenance window");
    }
  };

  // Handle form submission
  const handleSubmit = async () => {
    try {
      // Clear previous modal errors
      setModalError(null);

      // Validation
      if (!formData.start_time || !formData.end_time) {
        setModalError("Start time and end time are required");
        return;
      }

      if (!formData.message.trim()) {
        setModalError("Message is required");
        return;
      }

      const now = new Date();
      const startTime = new Date(formData.start_time);
      const endTime = new Date(formData.end_time);

      // Validate start time is not in the past (only for new windows or when changing start time)
      if (!editingWindow) {
        // Creating new window - start time must be in future
        if (startTime < now) {
          setModalError("Start time cannot be in the past");
          return;
        }
      } else {
        // Updating existing window - only validate if start time is being changed
        const existingStartTime = new Date(editingWindow.start_time);
        if (startTime.getTime() !== existingStartTime.getTime() && startTime < now) {
          setModalError("Start time cannot be in the past");
          return;
        }
      }

      // Validate end time is after start time
      if (endTime <= startTime) {
        setModalError("End time must be after start time");
        return;
      }

      const payload = {
        enabled: formData.enabled,
        start_time: convertToUTC(formData.start_time),
        end_time: convertToUTC(formData.end_time),
        message: formData.message.trim(),
      };

      let response;
      if (editingWindow) {
        // Update existing window
        response = await updateMaintenanceWindow(editingWindow.id, payload);
        if (response.type === "UPDATE_MAINTENANCE_WINDOW") {
          showNotification("success", "Maintenance window updated successfully");
        }
      } else {
        // Create new window
        response = await createMaintenanceWindow(payload);
        if (response.type === "CREATE_MAINTENANCE_WINDOW") {
          showNotification("success", "Maintenance window created successfully");
        }
      }

      if (response.type === "API_ERROR") {
        const errorMsg = response.payload?.response?.data?.error || "Operation failed";
        setModalError(errorMsg);
        return;
      }

      setIsModalOpen(false);
      setModalError(null);
      fetchWindows();
    } catch (error) {
      console.error("Error saving maintenance window:", error);
      setModalError("Failed to save maintenance window");
    }
  };

  // Handle delete
  const handleDelete = async (windowId) => {
    if (!window.confirm("Are you sure you want to delete this maintenance window?")) {
      return;
    }

    try {
      const response = await deleteMaintenanceWindow(windowId);
      if (response.type === "DELETE_MAINTENANCE_WINDOW") {
        showNotification("success", "Maintenance window deleted successfully");
        fetchWindows();
      } else if (response.type === "API_ERROR") {
        const errorMsg = response.payload?.response?.data?.error || "Delete failed";
        showNotification("error", errorMsg);
      }
    } catch (error) {
      console.error("Error deleting maintenance window:", error);
      showNotification("error", "Failed to delete maintenance window");
    }
  };

  // Filter windows based on search
  const filteredWindows = windows.filter((w) =>
    w.message.toLowerCase().includes(searchValue.toLowerCase())
  );

  const headers = [
    { key: "enabled", header: "Status", isSortable: true },
    { key: "start_time", header: "Start Time", isSortable: true },
    { key: "end_time", header: "End Time", isSortable: true },
    { key: "message", header: "Message", isSortable: true },
    { key: "actions", header: "Actions" },
  ];

  const rows = filteredWindows.map((window) => ({
    id: window.id,
    enabled: window.enabled ? "Enabled" : "Disabled",
    start_time: formatDate(window.start_time),
    end_time: formatDate(window.end_time),
    message: window.message,
    actions: window,
  }));

  return (
    <div className="maintenance-manager">
      <div className="maintenance-manager__header">
        <h2>Maintenance Windows</h2>
        <p>Manage scheduled maintenance notifications for users</p>
      </div>

      {notification && (
        <InlineNotification
          kind={notification.kind}
          title={notification.kind === "success" ? "Success" : "Error"}
          subtitle={notification.message}
          onCloseButtonClick={() => setNotification(null)}
          lowContrast
        />
      )}

      <DataTable rows={rows} headers={headers}>
        {({
          rows,
          headers,
          getHeaderProps,
          getRowProps,
          getTableProps,
          getTableContainerProps,
        }) => (
          <TableContainer
            title=""
            description=""
            {...getTableContainerProps()}
          >
            <TableToolbar>
              <TableToolbarContent>
                <TableToolbarSearch
                  persistent={true}
                  tabIndex={0}
                  onChange={(e) => setSearchValue(e.target.value)}
                  placeholder="Search"
                />
                <Button
                  renderIcon={Add}
                  onClick={handleCreate}
                  size="sm"
                >
                  Add Maintenance Window
                </Button>
              </TableToolbarContent>
            </TableToolbar>
            <Table {...getTableProps()}>
              <TableHead>
                <TableRow>
                  {headers.map((header) => (
                    <TableHeader key={header.key} {...getHeaderProps({ header })}>
                      {header.header}
                    </TableHeader>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={headers.length}>Loading...</TableCell>
                  </TableRow>
                ) : rows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={headers.length}>
                      No maintenance windows found
                    </TableCell>
                  </TableRow>
                ) : (
                  rows.map((row) => (
                    <TableRow key={row.id} {...getRowProps({ row })}>
                      {row.cells.map((cell) => {
                        if (cell.info.header === "actions") {
                          const window = cell.value;
                          return (
                            <TableCell key={cell.id}>
                              <OverflowMenu flipped>
                                <OverflowMenuItem
                                  itemText={window.enabled ? "Disable" : "Enable"}
                                  onClick={() => handleToggleEnabled(window)}
                                />
                                <OverflowMenuItem
                                  itemText="Edit"
                                  onClick={() => handleEdit(window)}
                                />
                                <OverflowMenuItem
                                  itemText="Delete"
                                  isDelete
                                  onClick={() => handleDelete(window.id)}
                                />
                              </OverflowMenu>
                            </TableCell>
                          );
                        }
                        return <TableCell key={cell.id}>{cell.value}</TableCell>;
                      })}
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DataTable>

      {!loading && totalItems > 0 && (
        <Pagination
          backwardText="Previous page"
          forwardText="Next page"
          itemsPerPageText="Items per page:"
          page={page}
          pageSize={pageSize}
          pageSizes={[10, 20, 30, 40, 50]}
          totalItems={totalItems}
          onChange={handlePaginationChange}
        />
      )}

      <Modal
        open={isModalOpen}
        onRequestClose={() => {
          setIsModalOpen(false);
          setModalError(null);
        }}
        onRequestSubmit={handleSubmit}
        modalHeading={editingWindow ? "Edit Maintenance Window" : "Create Maintenance Window"}
        primaryButtonText={editingWindow ? "Update" : "Create"}
        secondaryButtonText="Cancel"
        size="md"
      >
        <div className="maintenance-form">
          {modalError && (
            <InlineNotification
              kind="error"
              title="Error"
              subtitle={modalError}
              onCloseButtonClick={() => setModalError(null)}
              lowContrast
              style={{ marginBottom: "1rem" }}
            />
          )}

          <Toggle
            id="enabled-toggle"
            labelText="Enable this maintenance window"
            toggled={formData.enabled}
            onToggle={(checked) => setFormData({ ...formData, enabled: checked })}
          />

          <TextInput
            id="start-time"
            type="datetime-local"
            labelText="Start Time"
            value={formData.start_time}
            onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
            min={new Date().toISOString().slice(0, 16)}
            step="60"
            className="datetime-input"
            helperText="Select maintenance start date and time"
          />

          <TextInput
            id="end-time"
            type="datetime-local"
            labelText="End Time"
            value={formData.end_time}
            onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
            min={formData.start_time || new Date().toISOString().slice(0, 16)}
            step="60"
            className="datetime-input"
            helperText="Select maintenance end date and time"
          />

          <TextArea
            id="message"
            labelText="Notification Message"
            placeholder="Enter the maintenance notification message"
            value={formData.message}
            onChange={(e) => setFormData({ ...formData, message: e.target.value })}
            rows={4}
          />
        </div>
      </Modal>
    </div>
  );
};

export default MaintenanceManager;
