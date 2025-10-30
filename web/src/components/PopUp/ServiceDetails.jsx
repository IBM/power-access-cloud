// import axios from "axios";
import React from "react";
import { Modal } from "@carbon/react";
import { useNavigate } from "react-router-dom";

const ServiceDetails = ({pagename, selectRows, setActionProps }) => {
    let navigate = useNavigate();

   
  return (
    <Modal
      modalHeading="Service details"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        setActionProps("");
        navigate(pagename);
      }}
      open={true}
      primaryButtonText={"OK"}
      secondaryButtonText={"Cancel"}

    >
      
      {/* <p><strong>ID</strong>: {selectRows[0].id}</p>
        <p><strong>Name</strong>: {selectRows[0].name}</p> */}
        <p><strong>Display name</strong>: {selectRows[0].display_name}</p>
      
      <p><strong>Catalog name</strong>: {selectRows[0].catalog_name}</p>
        <p><strong>Expiry</strong>: {selectRows[0].expiry.split("T")[0]}</p>
      
        
        <p><strong>Access information</strong>: {selectRows[0].status.access_info==="..."?"Under process":"Access your VM at the external IP address: "+selectRows[0].status.access_info+". Use the SSH public key you added before you deployed this service."}</p>
        {selectRows[0].status.state==="PENDING EXTENSION"&&<p><strong>Requested extension date</strong>: {selectRows[0].status.extentiondate.split('T')[0]}<br/><br/><strong>Justification for extending</strong>: {selectRows[0].status.justification}</p>}
        
        <p><strong>Status</strong>: {selectRows[0].status.state}</p>
        <p><strong>Message</strong>: {selectRows[0].status.message}</p>
    </Modal>
  );
};

export default ServiceDetails;
