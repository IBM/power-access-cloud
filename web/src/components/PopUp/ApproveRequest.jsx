// import axios from "axios";
import React from "react";
import { approveRequest } from "../../services/request";
import { Modal } from "@carbon/react";

const ApproveRequest = ({selectRows,setActionProps})=> {
  const id = selectRows[0]?.id;

  const onSubmit = async () => {
    try {
      await approveRequest(id) // wait for the dispatch to complete
    } catch (error) {
      console.log(error);
    }
  };

  return (
    <Modal
      modalHeading="Exit group"
      danger={true}
      onRequestClose={()=>{
        setActionProps("");
      }}
      onRequestSubmit={()=>{
        onSubmit();
        setActionProps("");
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
    >
      <div>
         <div className="mb-3">
            <h4>Are you sure want to Exit from the Group!</h4>
          </div>
      </div>
    </Modal>
  );
}

export default ApproveRequest;
