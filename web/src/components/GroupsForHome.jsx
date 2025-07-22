import React, { useState, useEffect } from "react";
import { allGroups,allRequests } from "../services/request";
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
  DataTableSkeleton,
  OverflowMenu,
  OverflowMenuItem,
  Button,
  Tooltip
} from "@carbon/react";
import { TrashCan, Information } from "@carbon/icons-react";

import { flattenArrayOfObject } from "./commonUtils";

import UpgradeGroup from "./PopUp/UpgradeGroup";

import UserService from "../services/UserService";
import Notify from "./utils/Notify";

import DeleteRequest from './PopUp/DeleteRequest'
const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_DELETE_REQUEST="DELETE_REQUEST"

const new_request={
  key: BUTTON_REQUEST,
  label: "Request",
  kind: "ghost",
  standalone: true,
  hasIconOnly: true,
};

const delete_request={
  key: BUTTON_DELETE_REQUEST,
  label: "Request",
  kind: "ghost",
  standalone: true,
  hasIconOnly: true,
};

const headers = [
  {
    key: "name",
    header: "Group",
  },
  {
    key: "quota.cpu",
    header: "vCPU quota",
  },
  {
    key: "quota.memory",
    header: "Memory quota",
  },
  {
    key: "membership",
    header: "Status",
  },
  {
    key: "action",
    header: "Action",
  }
];

let selectRows = [];
const GroupsForHome = () => {
  
  const [rows, setRows] = useState([]);
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const isAdmin = UserService.isAdminUser();
  const [allGroupdata,setAllGroupsdata]=useState([])
  const [pendinggroups,setPendingGroups]=useState([])

  const filteredHeaders = isAdmin
    ? headers // Display all buttons for admin users
    : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users

  const fetchData = async () => {
    let data = [];
    let request=[];
    data = await allGroups();
    data?.payload.map((g) => 
      g.name=g.name.replace(/^./, g.name[0].toUpperCase())
    );
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
    request= await allRequests();
    setAllGroupsdata(data)
    const newResult = request.payload.filter((d)=>d.type==='GROUP'&& d.state==="NEW");
    setPendingGroups(newResult)
    const result=data.payload.filter((d)=>d.membership);

    
    newResult.forEach((p)=>{
     var pendingItem=data.payload.filter((d)=>d.name.toLowerCase()===p.group.group.toLowerCase())
     
     result.push(pendingItem[0])
    })

    
    setRows(result);
    setLoading(false);
  };

  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };


  useEffect(() => {
    fetchData();
    
  }, [actionProps]); 

  const displayData = flattenArrayOfObject(rows);
 
  
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
          <UpgradeGroup
          pagename=''
            currentGroupId={rows[0].id}
            allgroupdata={allGroupdata.payload}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_DELETE_REQUEST && (
          <DeleteRequest
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
    return (<div>Your group access is pending. Check back soon for status.
      </div>)
  }

  return (
    <>
      <Notify
        title={title}
        message={message}
        nkind={notifyKind}
        setTitle={setTitle}
      />
      {loading ? (
        renderSkeleton()
      ) : (
        <>
          {renderActionModals()}
          <DataTable
            rows={displayData}
            headers={filteredHeaders}
            isSortable
            radio
          >
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
                  
                <h4> My group
                  <Tooltip align="bottom-left" size="lg" label="Groups control resource allocation by assigning the maximum vCPU and memory available to you. By default, all new users are added to the Extra-small group which includes .5 vCPU and 8 GB of memory. If you require more CPU and memory, you can upgrade your group with a valid use case. You can only be a member of one group at a time.">
                    <Button className="sb-tooltip-trigger" kind="ghost" size="sm">
                            <Information />
                          </Button>
                    </Tooltip></h4>
                <TableContainer
                  
                  {...getTableContainerProps()}
                >
                  {((rows.length>0)&&<Table {...getTableProps()} style={{marginTop:"2rem"}}>
                    <TableHead>
                      <TableRow>
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
                          {row.cells.map((cell,i) => (cell.value &&
                            // <TableCell key={cell.id}>{cell.value}</TableCell>
                            ((i!==3)?<TableCell key={cell.id}>{cell.value}</TableCell>:<TableCell key={cell.id}>{(row.cells[i].value==="Active"&&"Approved")}{(row.cells[i].value==="Inactive"&&"Pending")}</TableCell>)
                          ))}
                          <TableCell >
                            {row.cells[3].value==="Inactive"&&  <OverflowMenu size="sm" flipped>
                              <OverflowMenuItem 
                              key={delete_request.key}
                              onClick={() => 
                                {
                                  selectRows=[];
                                  selectRows.push(pendinggroups[0]);
                                  setActionProps(delete_request)
                                }
                                }
                              itemText="Delete request" />
                            </OverflowMenu>}
                          </TableCell>
                        </TableRow>
                      ))}
                       {
                    
                  }
                    </TableBody>
                   
                  </Table>)}
                  
                </TableContainer>
                {(rows.length===0 && <div style={{backgroundColor:"#f4f4f4",padding:"1rem",marginTop:"3rem"}}>
                   {renderNoDataEmptyState()}
                  </div>
                  ) }
                  <Button kind="tertiary" disabled={rows.length===0} style={{float:"right",marginTop:"1rem"}} onClick={() => {if(pendinggroups.length>0){ alert('You already have a pending request, you cannot create a new request at the moment')} else{setActionProps(new_request)}}}>
                        Upgrade
                      </Button>
                  
                </div>
                </>
              );
            }}
          </DataTable>
        </>
      )}
    </>
  );
};
export default GroupsForHome;
