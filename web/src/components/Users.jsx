import React, { useState, useEffect, useRef } from "react";
import { allUsers } from "../services/request";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
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
} from "@carbon/react";

const headers = [
  {
    key: "id",
    header: "ID",
  },
  {
    key: "username",
    header: "User Name",
  },
  {
    key: "firstname",
    header: "First Name",
  },
  {
    key: "lastname",
    header: "Last Name",
  },
  {
    key: "email",
    header: "Email",
  },
];

const Users = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const isFirstRender = useRef(true);

  const fetchData = async () => {
    let data = await allUsers();
    setRows(data?.payload);
    setLoading(false);
  };

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    fetchData();
  }, [headers]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = clientSearchFilter(searchText, rows);

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

  if (loading) {
    renderSkeleton();
  }
  return (
    <>
      <DataTable rows={displayData} headers={headers} isSortable>
        {({
          rows,
          headers,
          getTableProps,
          getHeaderProps,
          getToolbarProps,
          getTableContainerProps,
        }) => {
          return (
            <TableContainer
              title={"User Details"}
              {...getTableContainerProps()}
            >
              <TableToolbar {...getToolbarProps()}>
                <TableToolbarSearch
                  persistent={true}
                  // tabIndex={batchActionProps.shouldShowBatchActions ? -1 : 0}
                  onChange={(onInputChange) => {
                    setSearchText(onInputChange.target.value);
                  }}
                  placeholder={"Search"}
                />
              </TableToolbar>
              <Table {...getTableProps()}>
                <TableHead>
                  <TableRow>
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

export default Users;
