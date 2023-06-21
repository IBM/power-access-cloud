import React, { useEffect,useState } from "react";
import { allRequests } from "../services/request";
import { APPROVE_REQUEST } from '../store/actionConstants';
import { MobileAdd } from "@carbon/icons-react";
import FooterPagination from "../utils/Pagination";
import { clientSearchFilter } from "../utils/Search";
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
} from '@carbon/react';
import ApproveRequest from "./PopUp/ApproveRequest";
const headers = [
  {
    key:"id",
    header: "ID"
  },
  {
    key:"group_id",
    header:"Group ID"
  },
  {
    key:"user_id",
    header:"User ID"
  },
  {
    key:"requester",
    header: "Requester"
  },
  {
    key: "created_at",
    header: "Created",
  },
  {
    key:"type",
    header: "Type"
  },
  {
    key: "state",
    header: "State"
  }
];

const TABLE_BUTTONS = [
  {
    key: APPROVE_REQUEST,
    label: ('Approve request'),
    kind: 'ghost',
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
  }
]
let selectRows = [];
const RequestList = () => {
  const [rows, setRows] = useState([]);
  const [searchText, setSearchText] = useState("");
  const [actionProps, setActionProps] = useState("");

  const fetchAllRequest = async ()=>{
    let data = await allRequests();
    setRows(data?.payload);
  }

  useEffect(() => {
    fetchAllRequest();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const selectionHandler = (rows=[])=>{
    console.log(rows);
    selectRows = rows;
  }

  const renderActionModals = ()=> {
    return (
      <React.Fragment>
        {actionProps?.key === APPROVE_REQUEST && (
          <ApproveRequest 
            selectRows={selectRows}
            setActionProps={setActionProps}
          />
        )}
      </React.Fragment>
    );
  }

  const displayData = clientSearchFilter(searchText, rows);
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
        {<FooterPagination displayData={rows} /> }
      </>
  );
};

export default RequestList;
