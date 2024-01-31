import React, { useState, useEffect } from "react";
import { Grid, Column } from "@carbon/react";
import { allGroups } from "../services/request";
import AddKey from "./PopUp/AddKey";
import "../styles/common.scss";
import {
  Button,
  Tile,
  InlineNotification,
  Tooltip
} from "@carbon/react";
import { MobileAdd, Information, CheckmarkFilled, WarningFilled} from "@carbon/icons-react";
import { getAllCatalogs } from "../services/request";
import DeployCatalog from "./PopUp/DeployCatalog";

import QuotaWarning from "./PopUp/QuotaWarning";
import Notify from "./utils/Notify";
import "../styles/registration.scss";

const BUTTON_REQUEST = "BUTTON_REQUEST";
const BUTTON_ADD= "BUTTON_ADD";
const deploy=  {
    key: BUTTON_REQUEST,
    label: "Deploy",
    kind: "ghost",
    icon: MobileAdd,
    standalone: true
  };
let lc=true;
const Catalogs = () => {

  const [rows, setRows] = useState([]);
  const [id, setId] = useState("");
  const [title, setTitle] = useState("");
  const [notifyKind, setNotifyKind] = useState("");
  const [message, setMessage] = useState("");
  const [actionProps, setActionProps] = useState("");
  const [cpu, setCPU] = useState(0);
  const [memory, setMemory] = useState(0);
  const fetchData = async () => {
    let data = await getAllCatalogs();
    
    
    data?.payload.sort((a, b) => {
      let fa = a.capacity.cpu,
        fb = b.capacity.cpu;
      if (fa < fb) {
        return -1;
      }
      if (fa > fb) {
        return 1;
      }
      return 0;
    });
    setRows(data?.payload);
    let data2 = await allGroups();
    const result= data2.payload.filter((d)=>d.membership)
    //alert(result[0].quota.cpu+ " "+ result[0].quota.memory);
    
    if(result.length>0){
      setCPU(result[0]?.quota.cpu);
      setMemory(result[0]?.quota.memory);
    }
    
  };
  const action={
    key: BUTTON_ADD,
      label: "Add key",
      kind: "ghost",
      icon: MobileAdd,
      standalone: true,
      hasIconOnly: true,
  }
  const handleResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : (setNotifyKind("success"));
  };
  useEffect(() => {
    fetchData();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps
  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_ADD && (
          <AddKey
            pagename=''
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
        {actionProps?.key === BUTTON_REQUEST && (
          <DeployCatalog
            selectRows={id}
            setActionProps={setActionProps}
            response={handleResponse}
          />
        )}
      </React.Fragment>
    );
  };
  return (
    <>
    <Grid fullWidth>
                <div className="page-banner">
                      <h1 className="landing-page__sub_heading banner-header">
                      Catalog
                      </h1>
                      <p className="banner-text">Explore the catalog of Power Access Cloud services and select the one that best suits your project needs. Note: You must <a style={{color:"blue", textDecoration:"underline", cursor:"pointer"}} onClick={() => setActionProps(action)} href="#/">add a SSH key</a> before deploying a service.</p>
                  </div>
                  </Grid>
    
      <p style={{marginLeft: "1rem"}}>Viewing {rows.length} item(s)</p>
      <InlineNotification style={{margin: "0 0 2rem 1rem"}}
      lowContrast={lc}
                        kind="info"
                        title="Note"
                        subtitle="Depending on your group access, not all catalog items will be available to you. If you need resources that exceed your group quotas, consider upgrading to a new group."
                        onClose={() => {
                            setTitle("");
                        }}
                    />
                    <Notify title={title} message={message} nkind={notifyKind} setTitle={setTitle} />
          <QuotaWarning />
          <Grid className="landing-page" fullWidth>
          {renderActionModals()}
          
      {rows.map((row) => (
        
    <Column key={row.id}
        lg={4}
        md={4}
        sm={2}
      ><Tile style={{paddingBottom:"50px", marginBottom:"50px", height:"85%", border:"1px grey solid"}} >
        {(row.status.ready)&&<Tooltip align="top" style={{float:"right"}} label="Ready">
<Button className="sb-tooltip-trigger" kind="ghost" size="sm">
<CheckmarkFilled  style={{fill:"#24A148"}} />
      </Button>
</Tooltip>}
{!(row.status.ready)&&<Tooltip align="top" style={{float:"right"}} label="Not ready">
<Button className="sb-tooltip-trigger" kind="ghost" size="sm">
<WarningFilled style={{fill:"#FA4D56"}} />
      </Button>
</Tooltip>}
        <img src={row.image_thumbnail_reference} width="15%" height="auto" alt="centos" /><br/>
      <strong><em>{row.name}</em></strong><br/><br/>
      
vCPU: {row.capacity.cpu}<br/>
Memory: {row.capacity.memory} GB<br/><br/>
{/* {row.description}
<br /><br /> */}
<span style={{float:"right"}} >
<Button  size="sm" kind="tertiary" disabled={(row.capacity.cpu>=cpu)||(row.capacity.memory>=memory)} onClick={() => {setId(row); setActionProps(deploy)}}>Deploy</Button>
{((row.capacity.cpu>=cpu)||(row.capacity.memory>=memory))&&<Tooltip align="top" label="You do not have enough resources available to deploy this service. Select a different service or upgrade your group to proceed.">
<Button className="sb-tooltip-trigger" kind="ghost" size="sm">
        <Information />
      </Button>
</Tooltip>}
</span>


   </Tile>
    </Column>
    ))}
      </Grid>
        </>
  );
};
export default Catalogs;