// import axios from "axios";
import React, { useState } from "react";
import { deployCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeployCatalog = ({ selectRows, setActionProps, response }) => {
  const [catalogName, setCatalogName] = useState("");
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");

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
      const { type, payload } = await deployCatalog({
        catalog_name: name,
        display_name: catalogName,
      }); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "The deployment of the catalog has failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "The catalog was deployed successfully.";
      }
    } catch (error) {
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/catalogs");
  };

  return (
    <Modal
      modalHeading="Deploy Catalog"
      danger={true}
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
          <label htmlFor="Name" className="form-label">
            Name
          </label>
          <input
            type={"text"}
            className="form-control"
            placeholder="Enter the display name for the service"
            name="name"
            value={catalogName}
            onChange={(e) => setCatalogName(e.target.value)}
          />
        </div>
        <div className="mb-3">
          <label htmlFor="Name" className="form-label">
            Catalog
          </label>
          <input
            type={"text"}
            className="form-control"
            placeholder="Enter your name"
            name="name"
            value={name}
            disabled
          />
        </div>
      </div>
    </Modal>
  );
}

export default DeployCatalog;
