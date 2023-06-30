import React, { useState, useEffect } from "react";
import { allKeys } from "../services/request";
import { MobileAdd, TrashCan } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import UserService from "../services/UserService";
import AddKey from "./PopUp/AddKey";
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
  TableSelectAll,
  DataTableSkeleton,
  InlineNotification,
} from "@carbon/react";
import DeleteKey from "./PopUp/DeleteKey";
const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";

const headers = [
  {
    key: "id",
    header: "ID",
    adminOnly: true,
  },
  {
    key: "user_id",
    header: "User ID",
    adminOnly: true,
  },
  {
    key: "name",
    header: "Name",
  },
  {
    key: "content",
    header: "Content",
  },
];

const TABLE_BUTTONS = [
  {
    key: BUTTON_DELETE,
    label: "Delete Key",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
  },
  {
    key: BUTTON_REQUEST,
    label: "Add Key",
    kind: "ghost",
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
  },
];
let selectRows = [];

const Keys = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [errorTitle, setErrorTitle] = useState("");
  const [errorMsg, setErrorMsg] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchData = async () => {
    let data = await allKeys();
    setRows(data?.payload);
    setLoading(false);
  };

  const handleErrorMessage = (title, message) => {
    setErrorTitle(title);
    setErrorMsg(message);
  };

  const selectionHandler = (rows = []) => {
    selectRows = rows;
  };

  useEffect(() => {
    fetchData();
  }, [isAdmin, headers, actionProps]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = clientSearchFilter(searchText, rows);

  const renderSkeleton = () => {
    const headerLabels = filteredHeaders?.map((x) => x?.header);
    return (
      <DataTableSkeleton
        columnCount={headerLabels?.length}
        compact={false}
        headers={headerLabels}
        rowCount={10}
        zebra={false}
      />
    );
  };
  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_REQUEST && (
          <AddKey
            selectRows={selectRows}
            setActionProps={setActionProps}
            onError={handleErrorMessage}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <DeleteKey
            selectRows={selectRows}
            setActionProps={setActionProps}
            onError={handleErrorMessage}
          />
        )}
      </React.Fragment>
    );
  };

  if (loading) {
    renderSkeleton();
  }
  return (
    <>
      {renderActionModals()}
      {errorMsg && (
        <InlineNotification
          title={errorTitle}
          subtitle={errorMsg}
          onClose={() => {
            setErrorMsg("");
          }}
        />
      )}
      <DataTable rows={displayData} headers={filteredHeaders} isSortable>
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
            <TableContainer title={"Key Details"} {...getTableContainerProps()}>
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
                {batchActionProps.batchActions.map((action) => {
                  return (
                    <TableBatchAction
                      key={action.key}
                      renderIcon={action.icon}
                      disabled={
                        !(selectRows.length === 1) &&
                        action.key !== BUTTON_REQUEST
                      }
                      onClick={() => setActionProps(action)}
                    >
                      {action.label}
                    </TableBatchAction>
                  );
                })}
              </TableToolbar>
              <Table {...getTableProps()}>
                <TableHead>
                  <TableRow>
                    <TableSelectAll {...getSelectionProps()} />
                    {headers.map((header) => (
                      <TableHeader
                        key={header.key}
                        {...getHeaderProps({ header })}
                      >
                        {header.header}
                      </TableHeader>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {rows.map((row) => (
                    <TableRow key={row.id}>
                      <TableSelectRow {...getSelectionProps({ row })} />
                      {row.cells.map((cell) => (
                        <TableCell key={cell.id}>{cell.value}</TableCell>
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
  );
};
export default Keys;
