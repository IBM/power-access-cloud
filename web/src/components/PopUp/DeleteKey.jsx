// import axios from "axios";
import React from "react";
import { deleteKeys } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteKey = ({ selectRows, pagename, setActionProps, response }) => {

  const id = selectRows[0]?.id;
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await deleteKeys({ id }); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Key delete failed";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Key deleted successfully";
      }
    } catch (error) {
      
    }
    response(title, message, errored)
    setActionProps("");
    navigate(pagename);
  };

  return (
    <Modal
      modalHeading="Delete key"
      danger={true}
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Delete"}
      secondaryButtonText={"Cancel"}
    >
      <div>
        <div className="mb-3">
          <h4>Are you sure you want to delete this key?</h4>
        </div>
      </div>
    </Modal>
  );
}

export default DeleteKey;
