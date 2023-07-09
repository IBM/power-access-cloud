import React, { useState, useEffect } from "react";
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
} from "@carbon/react";
import { CalendarAddAlt, TrashCan } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import { flattenArrayOfObject } from "./commonUtils";
import { getServices } from "../services/request";
import DeleteService from "./PopUp/DeleteService";
import ServiceExtend from "./PopUp/ServiceExtend";
import UserService from "../services/UserService";
import Notify from "./utils/Notify";

const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_EXTEND = "BUTTON_EXTEND";

const headers = [
  {
    key: "user_id",
    header: "User ID",
    adminOnly: true,
  },
  {
    key: "name",
    header: "Name",
    adminOnly: true,
  },
  {
    key: "display_name",
    header: "Display name",
  },
  {
    key: "catalog_name",
    header: "Catalog",
  },
  {
    key: "expiry",
    header: "Expiry",
  },
  {
    key: "status.state",
    header: "State",
  },
  {
    key: "status.message",
    header: "Message",
  },
  {
    key: "status.access_info",
    header: "Access Information",
  },
];

const TABLE_BUTTONS = [
  {
    key: BUTTON_EXTEND,
    label: "Change Expiry",
    kind: "ghost",
    icon: CalendarAddAlt,
    standalone: true,
    hasIconOnly: true,
  },
  {
    key: BUTTON_REQUEST,
    label: "Delete",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
    hasIconOnly: true,
  },
];

let selectRows = [];
const Services = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [title, setTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchData = async () => {
    let data = await getServices();
    // override the id field to be the name of the service to make it easier for the actions like expiry or delete
    setRows(data?.payload.map((row) => ({ ...row, id: row.name })));
    setLoading(false);
  };

  const selectionHandler = (rows = []) => {
    selectRows = rows;
  };

  useEffect(() => {
    fetchData();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = flattenArrayOfObject(
    clientSearchFilter(searchText, rows)
  );

  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };

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
          <DeleteService
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_EXTEND && (
          <ServiceExtend
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
      </React.Fragment>
    );
  };

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
                  title={"Service Details"}
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
                    {batchActionProps.batchActions.map((action) => {
                      return (
                        <TableBatchAction
                          key={action.key}
                          renderIcon={action.icon}
                          disabled={!(selectRows.length === 1)}
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
                          <TableHeader {...getHeaderProps({ header })}>
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
      )}
    </>
  );
};
export default Services;
