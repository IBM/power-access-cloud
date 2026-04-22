import React, { useEffect, useState } from "react";
import { allRequests } from "../services/request";
import {
  APPROVE_REQUEST,
  REQUEST_DETAILS,
  REJECT_REQUEST,
  DELETE_REQUEST,
} from "../store/actionConstants";
import {
  CheckmarkOutline,
  InformationSquare,
  MisuseOutline,
  TrashCan,
} from "@carbon/icons-react";
import FooterPagination from "../utils/Pagination";
import { clientSearchFilter } from "../utils/Search";
import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  TableToolbar,
  TableToolbarSearch,
  DataTableSkeleton,
  Dropdown,
  TableToolbarContent,
  OverflowMenu,
  OverflowMenuItem,
  Modal,
} from "@carbon/react";
import ApproveRequest from "./PopUp/ApproveRequest";
import RequestDetails from "./PopUp/RequestDetail";
import RejectRequest from "./PopUp/RejectRequest";
import DeleteRequest from "./PopUp/DeleteRequest";
import UserService from "../services/UserService";
import Notify from "./utils/Notify";

const headers = [
  {
    key: "id",
    header: "ID",
    adminOnly: true,
  },
  {
    key: "type",
    header: "Type",
  },
  {
    key: "username",
    header: "Username",
    adminOnly: true,
  },
  {
    key: "email",
    header: "Email",
    adminOnly: true,
  },
  {
    key: "created_at",
    header: "Created",
  },
  {
    key: "justification",
    header: "Justification",
  },
  {
    key: "state",
    header: "State",
  },
  {
    key: "comment",
    header: "Admin comments",
  },
  {
    key: "actions",
    header: "Actions",
  },
];

const TABLE_BUTTONS = [
  {
    key: REQUEST_DETAILS,
    label: "Details",
    kind: "ghost",
    icon: InformationSquare,
    standalone: true,
    hasIconOnly: true,
    adminOnly: false,
  },
  {
    key: DELETE_REQUEST,
    label: "Delete",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
    hasIconOnly: true,
    adminOnly: false,
  },
  {
    key: REJECT_REQUEST,
    label: "Reject",
    kind: "ghost",
    icon: MisuseOutline,
    standalone: true,
    hasIconOnly: true,
    adminOnly: true,
  },
  {
    key: APPROVE_REQUEST,
    label: "Approve",
    kind: "ghost",
    icon: CheckmarkOutline,
    standalone: true,
    hasIconOnly: true,
    adminOnly: true,
  },
];
let selectRows = [];

// Helper function to format date to user's local format
const formatDate = (dateString) => {
  if (!dateString) return "";
  try {
    const date = new Date(dateString);
    return date.toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  } catch (error) {
    return dateString; // Return original if parsing fails
  }
};

const RequestList = () => {
  const isAdmin = UserService.isAdminUser();
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [actionProps, setActionProps] = useState("");
  const [loading, setLoading] = useState(true);
  const [filterType, setFilterType] = useState("");
  const [filterState, setFilterState] = useState("");
  const [filterUsername, setFilterUsername] = useState("");
  const [confirmAction, setConfirmAction] = useState(null);

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchAllRequest = async () => {
    let data = await allRequests();
    // Format the created_at dates to user's local format
    const formattedData = data?.payload?.map(row => ({
      ...row,
      created_at: formatDate(row.created_at)
    }));
    setRows(formattedData);
    setLoading(false);
  };

  const filteredButtons = isAdmin
    ? TABLE_BUTTONS // Display all buttons for admin users
    : TABLE_BUTTONS.filter((button) => !button.adminOnly); // Filter out admin-only buttons for non-admin users

  useEffect(() => {
    fetchAllRequest();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
    // Refresh the table data after successful operations
    if (!errored) {
      fetchAllRequest();
    }
  };

  const handleActionClick = (action, row) => {
    // Find the original data row by matching the id
    const originalRow = rows.find(r => r.id === row.id);
    selectRows = [originalRow || row];
    
    // Reject action goes directly to its modal (which has comment field and confirmation)
    if (action.key === REJECT_REQUEST) {
      setActionProps(action);
    }
    // Show confirmation modal for Delete and Approve actions
    else if (action.key === DELETE_REQUEST || action.key === APPROVE_REQUEST) {
      setConfirmAction({ action, row: originalRow });
    } else {
      setActionProps(action);
    }
  };

  const handleConfirmAction = () => {
    if (confirmAction) {
      setActionProps(confirmAction.action);
      setConfirmAction(null);
    }
  };

  const getConfirmationMessage = () => {
    if (!confirmAction) return '';
    
    switch (confirmAction.action.key) {
      case DELETE_REQUEST:
        return 'Are you sure you want to delete this request? This action cannot be undone.';
      case REJECT_REQUEST:
        return 'Are you sure you want to reject this request?';
      case APPROVE_REQUEST:
        return 'Are you sure you want to approve this request?';
      default:
        return '';
    }
  };

  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === APPROVE_REQUEST && (
          <ApproveRequest
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === REQUEST_DETAILS && (
          <RequestDetails
            selectRows={selectRows}
            setActionProps={setActionProps}
          />
        )}
        {actionProps?.key === REJECT_REQUEST && (
          <RejectRequest
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === DELETE_REQUEST && (
          <DeleteRequest
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
      </React.Fragment>
    );
  };

  const renderSkeleton = () => {
    const headerLabels = filteredHeaders?.map((x) => x?.header);
    return (
      <DataTableSkeleton
        columnCount={headerLabels?.length}
        compact={false}
        headers={headerLabels}
        rowCount={3}
        zebra={false}
      />
    );
  };

  // Apply filters to the data
  const applyFilters = (data) => {
    let filtered = data;
    
    if (filterType) {
      filtered = filtered.filter(row => row.type === filterType);
    }
    
    if (filterState) {
      filtered = filtered.filter(row => row.state === filterState);
    }
    
    if (filterUsername) {
      filtered = filtered.filter(row => row.username === filterUsername);
    }
    
    return filtered;
  };

  // Get unique values for filter dropdowns
  const getUniqueTypes = () => {
    const types = [...new Set(rows.map(row => row.type).filter(Boolean))];
    return types.map(type => ({ id: type, label: type }));
  };

  const getUniqueStates = () => {
    const states = [...new Set(rows.map(row => row.state).filter(Boolean))];
    return states.map(state => ({ id: state, label: state }));
  };

  const getUniqueUsernames = () => {
    const usernames = [...new Set(rows.map(row => row.username).filter(Boolean))];
    return usernames.map(username => ({ id: username, label: username }));
  };

  const filteredData = applyFilters(rows);
  const displayData = clientSearchFilter(searchText, filteredData);
  return (
    <>
      <Notify title={title} message={message} nkind={notifyKind} setTitle={setTitle} />
      {loading ? (renderSkeleton()) : (
        <>
          {renderActionModals()}
          <DataTable rows={displayData} headers={filteredHeaders} isSortable>
            {({
              rows,
              headers,
              getTableProps,
              getHeaderProps,
              getRowProps,
              getToolbarProps,
              getTableContainerProps,
            }) => {
              return (
                <TableContainer
                  title={"Requests Detail"}
                  {...getTableContainerProps()}
                >
                  <TableToolbar {...getToolbarProps()}>
                    <TableToolbarSearch
                      persistent={true}
                      onChange={(onInputChange) => {
                        setSearchText(onInputChange.target.value);
                      }}
                      placeholder={"Search"}
                    />
                    <TableToolbarContent>
                      <Dropdown
                        id="filter-type"
                        titleText=""
                        label="Filter by Type"
                        items={[{ id: "", label: "All Types" }, ...getUniqueTypes()]}
                        selectedItem={filterType ? { id: filterType, label: filterType } : { id: "", label: "All Types" }}
                        onChange={({ selectedItem }) => setFilterType(selectedItem?.id || "")}
                        size="md"
                      />
                      <Dropdown
                        id="filter-state"
                        titleText=""
                        label="Filter by State"
                        items={[{ id: "", label: "All States" }, ...getUniqueStates()]}
                        selectedItem={filterState ? { id: filterState, label: filterState } : { id: "", label: "All States" }}
                        onChange={({ selectedItem }) => setFilterState(selectedItem?.id || "")}
                        size="md"
                      />
                      <Dropdown
                          id="filter-username"
                          titleText=""
                          label="Filter by Username"
                          items={[{ id: "", label: "All Users" }, ...getUniqueUsernames()]}
                          selectedItem={filterUsername ? { id: filterUsername, label: filterUsername } : { id: "", label: "All Users" }}
                          onChange={({ selectedItem }) => setFilterUsername(selectedItem?.id || "")}
                          size="md"
                      />
                    </TableToolbarContent>
                  </TableToolbar>
                  <Table {...getTableProps()}>
                    <TableHead>
                      <TableRow>
                        {headers.map((header) => (
                          <TableHeader {...getHeaderProps({ header })}>
                            {header.header}
                          </TableHeader>
                        ))}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {rows.map((row) => (
                        <TableRow key={row.id}>
                          {row.cells.map((cell) => (
                            <TableCell key={cell.id}>
                              {cell.info.header === "actions" ? (
                                <OverflowMenu size="sm" flipped>
                                  {filteredButtons.map((btn) => (
                                    <OverflowMenuItem
                                      key={btn.key}
                                      itemText={btn.label}
                                      onClick={() => handleActionClick(btn, row)}
                                    />
                                  ))}
                                </OverflowMenu>
                              ) : (
                                cell.value
                              )}
                            </TableCell>
                          ))}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              );
            }}
          </DataTable>
          {<FooterPagination displayData={rows} />}
        </>
      )}
      
      {/* Confirmation Modal */}
      {confirmAction && (
        <Modal
          open={true}
          danger={confirmAction.action.key === DELETE_REQUEST}
          modalHeading={`Confirm ${confirmAction.action.label}`}
          primaryButtonText="Confirm"
          secondaryButtonText="Cancel"
          onRequestSubmit={handleConfirmAction}
          onRequestClose={() => setConfirmAction(null)}
        >
          <p>{getConfirmationMessage()}</p>
        </Modal>
      )}
    </>
  );
};

export default RequestList;
