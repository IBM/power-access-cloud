import React,{useEffect,useState} from 'react';
import { getServices, allRequests } from "../services/request";
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
import { useNavigate } from "react-router-dom";
import { flattenArrayOfObject } from "./commonUtils";
import UserService from "../services/UserService";
import { CheckmarkFilled,Pending,InProgress, Information, Renew,ErrorFilled } from "@carbon/icons-react";
import DeleteService from "./PopUp/DeleteService";
import ServiceExtend from "./PopUp/ServiceExtend";
import ServiceDetails from './PopUp/ServiceDetails';
import Notify from "./utils/Notify";
const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_EXTEND = "BUTTON_EXTEND";
const BUTTON_DETAILS = "BUTTON_DETAILS";
const extend_action={
  key: BUTTON_EXTEND,
  label: "Change Expiry",
  kind: "ghost",
  standalone: true,
  hasIconOnly: true,
}
const delete_action={
  key: BUTTON_REQUEST,
  label: "Delete",
  kind: "ghost",
  standalone: true,
  hasIconOnly: true,
}
const details_action={
  key: BUTTON_DETAILS,
  label: "Details",
  kind: "ghost",
  standalone: true,
  hasIconOnly: true,
}
const headers = [
  {
    key: "display_name",
    header: "Service name",
  },
  {
    key: "expiry",
    header: "Expiration date",
  },
  {
    key: "status.state",
    header: "Status",
  },
  {
    key: "status.access_info",
    header: "Access information",
  },
  {
    key: "action",
    header: "Action",
  }
];
const ServicesForHome=({groups})=> {
  let navigate = useNavigate();
  const [servicesrows, setServicesRows] = useState([]);
  const isAdmin = UserService.isAdminUser();
  const [loading, setLoading] = useState(true);
  const [actionProps, setActionProps] = useState("");
  const [title, setTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [message, setMessage] = useState("");
  const result=groups.filter((d)=>d.membership)
  const [selectRow, setSelectRow] = useState([])
  const filteredHeaders = isAdmin
  ? headers // Display all buttons for admin users
  : headers.filter((header) => !header.adminOnly); // Filter out admin-only buttons for non-admin users
  const fetchServicesData = async () => {
    let data = await getServices();
    let request= await allRequests();
    const newResult = request.payload.filter((d)=>d.type==='SERVICE_EXPIRY'&& d.state==="NEW");
    
    newResult.forEach((result)=>{
      
      data?.payload.forEach((item)=>{
         
         
        if(item.name===result.service.name && item.status.state!=="EXPIRED"){
          item.status.state="PENDING EXTENSION"
          item.status.extentiondate=result.service.expiry
          item.status.justification=result.justification
        }
      })
    })
    
    data.payload.map((i)=>{
      if(i.status.access_info===""){
        i.status.access_info="..."
      }else{
        var begining=i.status.access_info.indexOf("ExternalIP:")
        var end=i.status.access_info.indexOf("use any ")
        i.status.access_info= i.status.access_info.slice(begining+12,end)
      }
    })
    setServicesRows(data?.payload);
    setLoading(false);
  };
  // useEffect(()=>{
  //   fetchServicesData();
  //   const intervalId = setInterval(() => {
  //     fetchServicesData(); // Fetch data every 1 minutes
  //   }, 60000);
  //   return () => clearInterval(intervalId);
  // },[actionProps])

  useEffect(()=>{
    fetchServicesData();
  },[actionProps])

  const displayData = flattenArrayOfObject(
    servicesrows
  );
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
  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
    if(!errored){
      setTimeout(() => {
        fetchServicesData();
      }, 3000);
      
    }
    
  };
  const renderNoDataEmptyState=()=>{
    return (<div>There are no services to display. After your group access is approved, you can add a service from the catalog.
      </div>)
  }
  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_DETAILS && (
          <ServiceDetails
            pagename=''
            selectRows={selectRow}
            setActionProps={setActionProps}
          />
        )}
        {actionProps?.key === BUTTON_REQUEST && (
          <DeleteService
            pagename=''
            selectRows={selectRow}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_EXTEND && (
          <ServiceExtend
            pagename=''
            selectRows={selectRow}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
      </React.Fragment>
    );
  };
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
                <Button size="md" kind="ghost" onClick={()=> fetchServicesData()} style={{float:"right"}} >
                        <Renew size="24"/>
                      </Button>
                      <h4> My services
                  <Tooltip align="bottom-left" size="lg" label="Review your service details including status, expiration date, and access information. Request additional services from the catalog.">
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
                            ((i!==2)?<TableCell key={cell.id}>{cell.value}</TableCell>:<TableCell key={cell.id}>{row.cells[i].value==="PENDING EXTENSION"&&<><svg  x="0px" y="0px"
                            viewBox="0 0 32 32"  width="18px" height="18px" fill="#FA4D56">
                         <path d="M28,6c0-1.1-0.9-2-2-2h-4V2h-2v2h-8V2h-2v2H6C4.9,4,4,4.9,4,6v20c0,1.1,0.9,2,2,2h19v-2H6V6h4v2h2V6h8v2h2V6h4v16h2V6z"/>
                         <g>
                           <polygon points="24.9,23.9 23.5,22.5 22.1,23.9 23.5,25.3 24.9,26.7 26.3,28.1 27.7,26.7 26.3,25.3 	"/>
                           <polygon points="23.5,28.1 22.1,29.5 23.5,30.9 24.9,29.5 26.3,28.1 27.7,26.7 26.3,25.3 24.9,26.7 	"/>
                         </g>
                         <g>
                           <circle cx="10.3" cy="17" r="2"/>
                           <circle cx="22.3" cy="17" r="2"/>
                           <circle cx="16.3" cy="17" r="2"/>
                         </g>
                         </svg> Pending Extension</>} {(row.cells[i].value==="CREATED"&&<> <CheckmarkFilled  style={{fill:"#24A148"}}/> Active</>)}{(row.cells[i].value==="NEW"&&<> <Pending style={{fill: "#FA4D56"}}/> Pending</>)}{(row.cells[i].value==="IN_PROGRESS"&&<> <InProgress style={{fill: "#F1C21B"}} /> Deploying</>)}{(row.cells[i].value==="EXPIRED"&&<> <ErrorFilled style={{fill: "#FA4D56"}} /> Expired</>)}</TableCell>)
                          ))}
                          <TableCell >
                            <OverflowMenu size="sm" flipped>
                              <OverflowMenuItem 
                              key={details_action.key}
                              onClick={() =>
                               {
                                const selectedrow=servicesrows.filter((d)=>
                                d.id===row.id
                                )
                                 setSelectRow(selectedrow)
                                 setActionProps(details_action)
                               } }
                              itemText="View details" />
                              {(row.cells[2].value!=="PENDING EXTENSION"&&row.cells[2].value!=="EXPIRED") && <OverflowMenuItem
                              key={extend_action.key}
                               onClick={() =>
                                {
                                  const selectedrow=displayData.filter((d)=>
                                d.id===row.id
                                )
                                 setSelectRow(selectedrow)
                                  setActionProps(extend_action)
                                } }
                                itemText="Extend service" />
                              }
                              <OverflowMenuItem
                               key={details_action.key}
                               onClick={() =>
                                {
                                  const selectedrow=displayData.filter((d)=>
                                d.id===row.id
                                )
                                 setSelectRow(selectedrow)
                                  setActionProps(delete_action)
                                }
                                } itemText="Delete service" />
                            </OverflowMenu>
                          </TableCell>
                        </TableRow>
                      ))}
                       {
                  }
                    </TableBody>
                  </Table>)}
                </TableContainer>
                {(rows.length===0 &&<div style={{backgroundColor:"#F4F4F4",padding:"1rem",marginTop:"3rem"}}>
        {(renderNoDataEmptyState()) }
        </div>)}
        <Button kind="tertiary" onClick={()=>navigate('/catalogs')} disabled={result.length===0} style={{float:"right",marginTop:"1rem"}} >
                        Go to catalog
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
}
export default ServicesForHome