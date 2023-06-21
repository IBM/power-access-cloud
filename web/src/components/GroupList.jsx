import React , { useState, useEffect } from "react";
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
  TableSelectAll,
  DataTableSkeleton
} from '@carbon/react';
import { MobileAdd, TrashCan } from "@carbon/icons-react";
import { clientSearchFilter } from "../utils/Search";
import FooterPagination from "../utils/Pagination";
import NewRequest from "./PopUp/NewRequest";
import ExitGroup from "./PopUp/ExitGroup";
const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";

const headers = [
  {
    key: 'id',
    header: 'ID',
  },
  {
    key: 'name',
    header: 'Groups',
  },
];

const TABLE_BUTTONS = [
  {
    key: BUTTON_REQUEST,
    label: ('Request group'),
    kind: 'ghost',
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
  },
  {
    key: BUTTON_DELETE,
    label: ('Exit group'),
    kind: 'ghost',
    icon: TrashCan,
    standalone: true,
  },
]
let selectRows = [];
const GroupList = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");

  const fetchData = async ()=>{
    let data = await allGroups();
    setRows(data?.payload);
    setLoading(false);
  }

  const selectionHandler = (rows=[])=>{
    console.log(rows);
    selectRows = rows;
  }

  useEffect(() => { 
    fetchData();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

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
  }
  const renderActionModals = ()=> {
    console.log("Test", actionProps);
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_REQUEST && (
          <NewRequest
            selectRows={selectRows}
            setActionProps={setActionProps}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <ExitGroup
            selectRows={selectRows}
            setActionProps={setActionProps}
          />
        )}
      </React.Fragment>
    );
  }

  if (loading){
    renderSkeleton();
  }
  return (  
      <>
      {renderActionModals()}
      <DataTable rows={displayData} headers={headers}>
      {({ rows, 
          headers, 
          getTableProps,
          getHeaderProps, 
          getRowProps,
          getBatchActionProps,
          getToolbarProps,
          getTableContainerProps,
          getSelectionProps,
          selectedRows }) => {
            const batchActionProps = getBatchActionProps({ batchActions: TABLE_BUTTONS });
            return (
              <TableContainer
                title={"Groups Detail"}
                {...getTableContainerProps()}>
                  {selectionHandler &&
                selectionHandler(selectedRows)}
                <TableToolbar {...getToolbarProps()}>
                  <TableToolbarSearch
                    persistent="true"
                    tabIndex={batchActionProps.shouldShowBatchActions ? -1 : 0}
                    onChange={onInputChange => {
                      setSearchText(onInputChange.target.value);
                    }}
                    placeholder={('Search')}
                  />
                  {batchActionProps.batchActions.map((action) => {
                    console.log({action})
                      return <TableBatchAction
                        renderIcon={action.icon}
                        disabled={!(selectRows.length === 1)}
                        onClick={()=>setActionProps(action)}
                      >
                      {action.label}
                    </TableBatchAction>
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
                      <TableRow>
                        <TableSelectRow {...getSelectionProps({row})} />
                        {row.cells.map((cell) => (
                          <TableCell key={cell.id}>{cell.value}</TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )
          }
        }
      </DataTable>
      { <FooterPagination displayData={rows} /> }
      </>
  );
};
export default GroupList;
