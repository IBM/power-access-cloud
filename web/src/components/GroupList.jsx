import React, { useState, useEffect } from "react";
import { allGroups } from "../services/request";

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
import { MobileAdd, TrashCan } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import { flattenArrayOfObject } from "./commonUtils";
import NewRequest from "./PopUp/NewRequest";
import ExitGroup from "./PopUp/ExitGroup";
import UserService from "../services/UserService";
import Notify from "./utils/Notify";

const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";

const headers = [
  {
    key: "id",
    header: "ID",
    adminOnly: true,
  },
  {
    key: "name",
    header: "Groups",
  },
  {
    key: "quota.cpu",
    header: "Quota CPU",
  },
  {
    key: "quota.memory",
    header: "Quota Memory(GB)",
  },
  {
    key: "membership",
    header: "Membership",
  },
];

const TABLE_BUTTONS = [
  {
    key: BUTTON_REQUEST,
    label: "Request",
    kind: "ghost",
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
  },
  {
    key: BUTTON_DELETE,
    label: "Exit",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
  },
];
let selectRows = [];
const GroupList = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

    const fetchData = async () => {
      let data = [];
      data = await allGroups();
        data?.payload.sort((a, b) => {
          let fa = a.quota.cpu,
            fb = b.quota.cpu;
          if (fa < fb) {
            return -1;
          }
          if (fa > fb) {
            return 1;
          }
          return 0;
        });
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
          <NewRequest
            selectRows={selectRows}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <ExitGroup
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
          <DataTable rows={displayData} headers={filteredHeaders} isSortable radio>
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
                  title={"Groups Detail"}
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
                        <TableHeader />
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
export default GroupList;
