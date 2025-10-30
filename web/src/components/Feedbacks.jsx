import { useState, useEffect } from "react";
import { Pagination } from "@carbon/react";
import { getFeedbacks } from "../services/request";
import { clientSearchFilter } from "../utils/Search";
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
    key: "user_id",
    header: "User ID",
  },
  {
    key: "rating",
    header: "Rating",
  },
  {
    key: "comment",
    header: "Comments",
  },
  {
    key: "created_at",
    header: "Created At",
  },
];

const Feedbacks = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [totalItems, setTotalItems] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [error, setError] = useState(null);

  const changePaginationState = async (e) => {
    setPage(e.page);
    setPageSize(e.pageSize);
    setLoading(true);
    let data = await getFeedbacks(e.page, e.pageSize);
    setTotalItems(data?.payload.total_items);
    setRows(data?.payload.feedbacks);
    setLoading(false);
  };

  useEffect(() => {
    const fetchData = async () => {
      const data = await getFeedbacks(page, pageSize);

      data.type === "API_ERROR"
        ? (() => {
            setError("Failed to fetch data");
            setLoading(false);
          })()
        : (() => {
            setTotalItems(data?.payload.total_items);
            setRows(data?.payload.feedbacks);
            setLoading(false);
          })();
    };
    fetchData();
  }, [headers]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData = () => {
    return flattenArrayOfObject(clientSearchFilter(searchText, rows));
  };

  const renderSkeleton = () => {
    const headerLabels = headers?.map((x) => x?.header);
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

  return (
    <>
      {loading || error ? (
        renderSkeleton()
      ) : (
        <>
          <DataTable rows={displayData()} headers={headers} isSortable>
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
                  title={"Feedback Dashboard"}
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
            pageSizes={[5, 10, 20, 30, 40, 50]}
            size="md"
            totalItems={totalItems}
          />
        </>
      )}
    </>
  );
};

export default Feedbacks;
