// import axios from "axios";
import React, { useState } from "react";
import { deleteServices } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteService = ({pagename, selectRows, setActionProps, response }) => {
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Delete");
  const getCellValue = (data, key) => {
    if (!data) return '';
    
    // Admin Persona (Carbon Table format)
    if (data.cells) {
      const cell = data.cells.find(c => c.id.endsWith(`:${key}`));
      return cell ? cell.value : '';
    }

    // User Persona (Standard Object format)
    return data[key] || '';
  };
  const name = getCellValue(selectRows[0], 'name');
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await deleteServices(name); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Service deletion failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Service deleted successfully.";
      }
    } catch (error) {
      
    }
    response(title, message, errored)
    setActionProps("");
    navigate(pagename);
  };

  return (
    <Modal
      modalHeading="Delete service"
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
      primaryButtonDisabled={primaryButtonDisabled}
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
