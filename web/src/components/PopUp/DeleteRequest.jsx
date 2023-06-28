// import axios from "axios";
import React from "react";
import { deleteRequest } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteRequest = ({ selectRows, setActionProps, onError }) => {
  let id = "";
  selectRows[0].cells.forEach((item) => {
    if (item.id.split(":")[1] === "id") {
      id = item?.value;
    }
  });
  let navigate = useNavigate();

  const onSubmit = async () => {
    try {
      const {type, payload} = await deleteRequest(id); // wait for the dispatch to complete
      if (type==="API_ERROR"){
        const errorTitle = "Request deletion failed"
        const errorMsg = payload.response.data.error;
        onError(errorTitle, errorMsg);
      }
    } catch (error) {
      console.log(error);
    }
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
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Delete"}
      secondaryButtonText={"Cancel"}
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
