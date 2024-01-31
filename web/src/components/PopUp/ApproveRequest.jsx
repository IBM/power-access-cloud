// import axios from "axios";
import React, { useState } from "react";
import { approveRequest } from "../../services/request";
import { Modal } from "@carbon/react";
import { useNavigate } from "react-router-dom";

const ApproveRequest = ({ selectRows, setActionProps, response }) => {
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");
  const id = selectRows[0]?.id;
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await approveRequest(id); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Approval of request failed";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Approved request successfully.";
      }
    } catch (error) {
      
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/requests");
  };

  return (
    <Modal
      modalHeading="Approve Request"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        setPrimaryButtonDisabled(true);
        setPrimaryButtonText("Submitting...")
        onSubmit();
      }}
      open={true}
      primaryButtonText={primaryButtonText}
      secondaryButtonText={"Cancel"}
      primaryButtonDisabled={primaryButtonDisabled}
    >
      <div>
        <div className="mb-3">
          <h4>Are you sure want to Approve this request!</h4>
        </div>
      </div>
    </Modal>
  );
};

export default ApproveRequest;
