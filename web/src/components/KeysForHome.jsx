import React, { useState, useEffect } from "react";
import { allKeys } from "../services/request";
import { MobileAdd, TrashCan, Information } from "@carbon/icons-react";
import UserService from "../services/UserService";
import AddKey from "./PopUp/AddKey";
//import { NoDataEmptyState } from "@carbon/ibm-products";
import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  Tooltip,
  DataTableSkeleton,
  OverflowMenu,
  OverflowMenuItem,
  Modal,
  Button,
} from "@carbon/react";
import "../styles/keysforhome.scss";
import DeleteKey from "./PopUp/DeleteKey";
import Notify from "./utils/Notify";

const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE = "BUTTON_DELETE";

const headers = [
  // {
  //   key: "id",
  //   header: "ID",
  //   adminOnly: true,
  // },
  // {
  //   key: "user_id",
  //   header: "User ID",
  //   adminOnly: true,
  // },
  {
    key: "name",
    header: "Name",
  },
  {
    key: "content",
    header: "Action",
  },
  
];

const action={
  key: BUTTON_REQUEST,
    label: "Add key",
    kind: "ghost",
    icon: MobileAdd,
    standalone: true,
    hasIconOnly: true,
}

const delete_action={
  key: BUTTON_DELETE,
    label: "Delete Key",
    kind: "ghost",
    icon: TrashCan,
    standalone: true,
}
let selectRows = [];

const KeysForHome = () => {
  const [rows, setRows] = useState([]);
  const [errorTitle, setErrorTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [errorMsg, setErrorMsg] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();

  const [open, setOpen] = useState(false);
  const [keyname, setKeyName] = useState("");
  const [keyvalue, setKeyValue] = useState("");
  const [iscoppied, setIscoppied] = useState(true);

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchData = async () => {
    let data = await allKeys();
    setRows(data?.payload);
    setLoading(false);
  };

  const handleResponse = (title, message, errored) => {
    setErrorTitle(title);
    setErrorMsg(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };

  useEffect(() => {
    fetchData();
  }, [isAdmin, headers, actionProps]); // eslint-disable-line react-hooks/exhaustive-deps

  const displayData =rows;

  const renderSkeleton = () => {
    const headerLabels = filteredHeaders?.map((x) => x?.header);
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
  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_REQUEST && (
          <AddKey
            pagename=''
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_DELETE && (
          <DeleteKey
            selectRows={selectRows}
            pagename=''
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
      </React.Fragment>
    );
  };

const renderNoDataEmptyState=()=>{
  return (<div>There are no keys to display. PAC requires a SSH public key to deploy services. Click Add key to proceed.</div>)
}
  return (
    <>
    <Modal
        modalHeading="Key details"
        secondaryButtonText="Cancel"
        primaryButtonText="Copy"
        open={open}
        onRequestClose={() => {
          setOpen(false);
          setIscoppied(true);
          setKeyName("");
          setKeyValue("");
        }}
        onRequestSubmit={()=>{
          setIscoppied(false);
          navigator.clipboard.writeText(keyvalue);
          
        }}
        
      >
        <p><strong>Key name</strong>: {keyname}</p>
        <p><strong>Key Value</strong>: <span className={`${iscoppied ? "" : "highlight"}`}>{keyvalue}</span></p>
      </Modal>
      <Notify title={errorTitle} message={errorMsg} nkind={notifyKind} setTitle={setErrorTitle} />
      {loading ? (renderSkeleton()) : (
        <>
          {renderActionModals()}
          <DataTable rows={displayData} headers={filteredHeaders} isSortable>
            {({
              rows,
              headers,
              getTableProps,
              getHeaderProps,
              getTableContainerProps,
              
            }) => {
              
              return (
                <>
                <div style={{padding:"1rem", border: "1px solid #E4E5E6",minHeight:"22rem",overflow:"hidden"}}>
                  <h4> My keys <Tooltip align="bottom-left" size="lg" label="Secure authentication to your services requires the use of keys. Multiple keys can be created using Add Key. Click on View Details to learn more about a specific key, and you can also delete any unused keys.">
                    <Button className="sb-tooltip-trigger" kind="ghost" size="sm">
                            <Information />
                          </Button>
                    </Tooltip></h4>
                <TableContainer {...getTableContainerProps()}>
                   
                  {(rows.length>0&&<Table {...getTableProps()} style={{marginTop:"2rem"}}>
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
                          
                          
                        <TableCell key={row.cells[0].id}>{row.cells[0].value}</TableCell>
                        

                          <TableCell>
                            <OverflowMenu size="sm" flipped>
                              <OverflowMenuItem
                                onClick={() => {
                                  setOpen(true);

                                  setKeyName(row.cells[0].value);
                                  setKeyValue(row.cells[1].value);
                                }}
                                itemText="View details"
                              />
                              
                              <OverflowMenuItem
                              key={delete_action.key}
                              renderIcon={delete_action.icon}
                              
                              onClick={() => 
                                {
                                  selectRows=[];
                                  selectRows.push(row);
                                  setActionProps(delete_action)
                                }
                                }
                              itemText="Delete key" />
                            </OverflowMenu>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>)}
                </TableContainer>
                {(rows.length===0 && <div style={{backgroundColor:"#f4f4f4",padding:"1rem",marginTop:"2rem"}}>
                    {(renderNoDataEmptyState())}
                  
                  </div>
                  ) }
                  
                  <Button kind="tertiary" style={{float:"right",marginTop:"1rem",marginBottom:"1rem",}}
                          key={action.key}
                          disabled={
                            !(selectRows.length === 1) &&
                            action.key !== BUTTON_REQUEST
                          }
                          onClick={() => setActionProps(action)}
                        >
                          {action.label}
                        </Button>
                  
                  
                </div>
                </>
              );
            }}
          </DataTable>
          
          <div>
          
            <br />
          
          </div>
          
          
        </>
      )}
    </>
  )
}
export default KeysForHome;
