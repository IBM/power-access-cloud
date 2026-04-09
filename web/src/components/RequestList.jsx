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
  TableBatchAction,
  TableSelectRow,
  TableToolbarSearch,
  DataTableSkeleton,
  Dropdown,
  TableToolbarContent,
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
  };

  const selectionHandler = (rows = []) => {
    selectRows = rows;
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
          <DataTable rows={displayData} headers={filteredHeaders} radio isSortable>
            {({
              rows,
              headers,
              getTableProps,
              getHeaderProps,
              getRowProps,
              getBatchActionProps,
              getToolbarProps,
              getTableContainerProps,
              getSelectionProps,
              selectedRows,
            }) => {
              const batchActionProps = getBatchActionProps({
                batchActions: TABLE_BUTTONS,
              });
              return (
                <TableContainer
                  title={"Requests Detail"}
                  {...getTableContainerProps()}
                >
                  {selectionHandler && selectionHandler(selectedRows)}
                  <TableToolbar {...getToolbarProps()}>
                    <TableToolbarSearch
                      persistent={true}
                      tabIndex={batchActionProps.shouldShowBatchActions ? -1 : 0}
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
                    {batchActionProps.batchActions.map((action) => {
                      return filteredButtons.map((btn) => {
                        if (btn.key === action.key) {
                          return (
                            <TableBatchAction
                              renderIcon={btn.icon}
                              disabled={!(selectRows.length === 1)}
                              onClick={() => setActionProps(btn)}
                              key={btn.key} // Add a unique key for each rendered component
                            >
                              {btn.label}
                            </TableBatchAction>
                          );
                        }
                        return null;
                      });
                    })}
                  </TableToolbar>
                  <Table {...getTableProps()}>
                    <TableHead>
                      <TableRow>
                        <th></th>
                        {headers.map((header) => (
                          <TableHeader {...getHeaderProps({ header })}>
                            {header.header}
                          </TableHeader>
                        ))}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {rows.map((row) => {
                        const selectionProps = getSelectionProps({ row });
                        return (
                          <TableRow 
                            key={row.id}
                            onClick={(e) => {
                              // Don't trigger if clicking on the radio button itself
                              if (e.target.type === 'radio') return;
                              
                              // Find and click the radio input in this row
                              const radioInput = e.currentTarget.querySelector('input[type="radio"]');
                              if (radioInput) {
                                radioInput.click();
                              }
                            }}
                            style={{ cursor: 'pointer' }}
                          >
                            <TableSelectRow {...selectionProps} />
                            {row.cells.map((cell) => (
                              <TableCell key={cell.id}>{cell.value}</TableCell>
                            ))}
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              );
            }}
          </DataTable>
          {<FooterPagination displayData={rows} />}
        </>
      )}
    </>
  );
};

export default RequestList;
