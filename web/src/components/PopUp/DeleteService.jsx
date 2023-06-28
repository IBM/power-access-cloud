// import axios from "axios";
import React from "react";
import { deleteServices } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteService = ({ selectRows, setActionProps, onError }) => {
  let name = "";
  selectRows[0].cells.forEach((item) => {
    if (item.id.split(":")[1] === "name") {
      name = item?.value;
    }
  });
  let navigate = useNavigate();

  const onSubmit = async () => {
    try {
      const {type, payload} = await deleteServices(name); // wait for the dispatch to complete
      if (type==="API_ERROR"){
        const errorTitle = "Service deletion failed"
        const errorMsg = payload.response.data.error;
        onError(errorTitle, errorMsg);
      }
    } catch (error) {
      console.log(error);
    }
    setActionProps("");
    navigate("/services");
  };

  return (
    <Modal
      modalHeading="Delete service"
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
          <h4>Are you sure want to delete this Service!</h4>
        </div>
      </div>
    </Modal>
  );
};

export default DeleteService;
