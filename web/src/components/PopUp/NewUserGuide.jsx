// import axios from "axios";
import React from "react";
import { Modal } from "@carbon/react";


const NewUserGuide = ({ setActionProps }) => {

  return (
    
    <Modal
      modalHeading="User Guide"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        alert('hi')
      }}
      open={true}
      primaryButtonText={"Ok"}
      secondaryButtonText={"Cancel"}
    >
      <div>
      <div className="mb-3">
          <h6>Step: 1</h6>
          Please add your public key in "My Keys" Section which will be used to login into the Power instance.
          <br />
          <h6>Step 2</h6>
          Once group access is approved, you can directly go to the "Catalog" section and choose one of the enabled catalogs as per your requirement. Click deploy.
           <br />
          <h6>Step 3: </h6>
          Enter details for your service (Power instance) and submit. Within few secs, instance will be available to access under "My Services" section.
      </div>
      </div>
    </Modal>
    
  );
}
export default NewUserGuide;
