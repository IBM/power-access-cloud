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
import { MobileAdd, TrashCan, AlarmSubtract } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import { flattenArrayOfObject } from "./commonUtils";
import { getAllCatalogs } from "../services/request";
import DeployCatalog from "./PopUp/DeployCatalog";
import DeleteCatalog from "./PopUp/DeleteCatalog";
import RetireCatalog from "./PopUp/RetireCatalog";
import UserService from "../services/UserService";
import Notify from "./utils/Notify";

const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";
const BUTTON_RETIRE = "BUTTON_RETIRE";

const headers = [
  {
    key: "name",
    header: "Name",
  },
  {
    key: "type",
    header: "Type",
  },
  {
    key: "description",
    header: "Description",
  },
  {
    key: "status.ready",
    header: "Status",
  },
  {
    key: "retired",
    header: "Retired",
  },
  {
    key: "status.message",
    header: "Message",
  },
];

const TABLE_BUTTONS = [
  {
    key: BUTTON_RETIRE,
    label: "Retire",
    kind: "ghost",
    icon: AlarmSubtract,
    standalone: true,
    hasIconOnly: true,
    adminOnly: true,
  },
  {
    key: BUTTON_DELETE,
    label: "Delete",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
    hasIconOnly: true,
    adminOnly: true,
  },
  {
    key: BUTTON_REQUEST,
    label: "Deploy",
    kind: "ghost",
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
    adminOnly: false,
  },
];

let selectRows = [];
const CatalogsAdmin = () => {
  const isAdmin = UserService.isAdminUser();
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [title, setTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");

  const filteredButtons = isAdmin
    ? TABLE_BUTTONS // Display all buttons for admin users
    : TABLE_BUTTONS.filter((button) => !button.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchData = async () => {
    let data = await getAllCatalogs();
    setRows(data?.payload);
    setLoading(false);
  };

  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
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
  const renderSkeleton = () => {
    const headerLabels = headers?.map((x) => x?.header);
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
        {actionProps?.key === BUTTON_RETIRE && (
          <RetireCatalog
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <DeleteCatalog
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_REQUEST && (
          <DeployCatalog
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
          <DataTable rows={displayData} headers={headers} isSortable>
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
                  title={"Catalog Detail"}
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
export default CatalogsAdmin;
