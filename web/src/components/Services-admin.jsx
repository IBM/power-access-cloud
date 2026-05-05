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
  DataTableSkeleton,
} from "@carbon/react";
import { CalendarAddAlt, TrashCan } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import { flattenArrayOfObject } from "./commonUtils";
import { getServices } from "../services/request";
import DeleteService from "./PopUp/DeleteService";
import AdminServiceExtend from "./PopUp/AdminServiceExtend";
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
    key: "username",
    header: "Username",
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

const ServicesAdmin = () => {
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
            selectRows={actionProps.selectRows || []}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_EXTEND && (
          <AdminServiceExtend
            selectRows={actionProps.selectRows || []}
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
          <DataTable rows={displayData} headers={filteredHeaders} isSortable radio>
            {({
              rows: tableRows,
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
                          disabled={!(selectedRows.length === 1)}
                          onClick={() => {
                            const selectedServiceRows = tableRows
                              .filter((tableRow) =>
                                selectedRows.some((selectedRow) => selectedRow.id === tableRow.id)
                              )
                              .map((tableRow) => rows.find((service) => service.id === tableRow.id))
                              .filter(Boolean);

                            setActionProps({
                              ...action,
                              selectRows: selectedServiceRows,
                            });
                          }}
                        >
                          {action.label}
                        </TableBatchAction>
                      );
                    })}
                  </TableToolbar>
                  <Table {...getTableProps()}>
                    <TableHead>
                      <TableRow>
                        <TableHeader />
                        {headers.map((header) => (
                          <TableHeader {...getHeaderProps({ header })}>
                            {header.header}
                          </TableHeader>
                        ))}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {tableRows.map((row) => (
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
export default ServicesAdmin;