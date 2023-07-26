import React, { useState, useEffect } from "react";
import { getEvents } from "../services/request";
import { clientSearchFilter } from "../utils/Search";
import { Pagination } from '@carbon/react';
import { flattenArrayOfObject } from "./commonUtils";
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
    key: "type",
    header: "Type",
  },
  {
    key: "created_at",
    header: "Created At",
  },
  {
    key: "user_id",
    header: "User ID",
  },
  {
    key: "log.level",
    header: "Log Level",
  },
  {
    key: "log.message",
    header: "Log Message",
  },
];

const Events = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [totalItems, setTotalItems] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const fetchData = async () => {
    let data = await getEvents(page, pageSize);
    console.log(data?.payload);
    setTotalItems(data?.payload.total_items);
    setRows(data?.payload.events);
    setLoading(false);
  };

  const changePaginationState = async (e) => {
    setPage(e.page);
    setPageSize(e.pageSize);
    setLoading(true);
    let data = await getEvents(e.page, e.pageSize);
    setTotalItems(data?.payload.total_items);
    setRows(data?.payload.events);
    setLoading(false);
  };

  useEffect(() => {
    fetchData();
  }, [headers]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = flattenArrayOfObject(clientSearchFilter(searchText, rows));

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

  return (
    <>
      {loading ? (renderSkeleton()) : (
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
                  title={"Event Details"}
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
          <Pagination
            onChange={changePaginationState}
            backwardText="Previous page"
            forwardText="Next page"
            itemsPerPageText="Items per page:"
            page={page}
            pageNumberText="Page Number"
            pageSize={pageSize}
            pageSizes={[
              5,
              10,
              20,
              30,
              40,
              50
            ]}
            size="md"
            totalItems={totalItems}
          />
        </>
      )}
    </>
  );
};

export default Events;
