// import axios from "axios";
import React from "react";
import { deleteCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteCatalog = ({ selectRows, setActionProps, response }) => {
  let name = "";
  selectRows[0].cells.forEach((item) => {
    if (item.id.split(":")[1] === "name") {
      name = item?.value;
    }
  });
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await deleteCatalog(name); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "The deletion of the catalog has failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "The deletion of the catalog was successful.";
      }
    } catch (error) {
      
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/catalogs");
  };

  return (
    <Modal
      modalHeading="Delete Catalog"
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
          <h4>Are you sure want to delete this Catalog!</h4>
        </div>
      </div>
    </Modal>
  );
}

export default DeleteCatalog;
