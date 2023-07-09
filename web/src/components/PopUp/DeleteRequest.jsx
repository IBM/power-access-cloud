// import axios from "axios";
import React, { useState } from "react";
import { deleteRequest } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteRequest = ({ selectRows, setActionProps, response }) => {
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Delete");
  const id = selectRows[0]?.id;
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await deleteRequest(id); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Request deletion failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Deleted request successfully.";
      }
    } catch (error) {
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/requests");
  };

  return (
    <Modal
      modalHeading="Delete Request"
      danger={true}
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        setPrimaryButtonDisabled(true);
        setPrimaryButtonText("Deleting...")
        onSubmit();
      }}
      open={true}
      primaryButtonText={primaryButtonText}
      secondaryButtonText={"Cancel"}
      primaryButtonDisabled={primaryButtonDisabled}
    >
      <div>
        <div className="mb-3">
          <h4>Are you sure want to delete this Request?</h4>
        </div>
      </div>
    </Modal>
  );
};

export default DeleteRequest;
