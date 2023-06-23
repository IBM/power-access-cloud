import React , { useState, useEffect, useRef } from "react";
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
    DataTableSkeleton
  } from '@carbon/react';
import DeleteKey from "./PopUp/DeleteKey";
const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";

const headers = [
  {
    key: 'id',
    header: 'ID',
  },
  {
    key: 'name',
    header: 'Name',
  },
  {
    key: 'content',
    header: "Content"
  }
];

const TABLE_BUTTONS = [
    {
        key: BUTTON_DELETE,
        label: ('Delete Key'),
        kind: 'ghost',
        icon: TrashCan,
        standalone: true,
    },
    {
        key: BUTTON_REQUEST,
        label: ('Add Key'),
        kind: 'ghost',
        icon: MobileAdd,
        standalone: true,
        hasIconOnly: true,
    }
]
let selectRows = [];

const Keys = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();
  const isFirstRender = useRef(true);
  
  const fetchData = async ()=>{
    let data = await allKeys();
    setRows(data?.payload);
    setLoading(false);
  }
  const selectionHandler = (rows=[])=>{
    console.log(rows);
    selectRows = rows;
  }

  useEffect(() => { 
    if (isFirstRender.current) {
        isFirstRender.current = false;
        return;
    }
    if (isAdmin){
        headers.splice(2, 0, {
            key: 'user_id',
            header: "User ID"
        },);
    }
    fetchData();
  }, [isAdmin,headers,actionProps]); // eslint-disable-line react-hooks/exhaustive-deps

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
          <AddKey
            selectRows={selectRows}
            setActionProps={setActionProps}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <DeleteKey
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
                title={"Key Details"}
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
                        disabled={!(selectRows.length === 1) && action.key !== BUTTON_REQUEST}
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
export default Keys;
