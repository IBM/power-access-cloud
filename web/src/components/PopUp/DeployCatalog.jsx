// import axios from "axios";
import React, { useState } from "react";
import { deployCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeployCatalog = ({selectRows, setActionProps, onError })=> {
    const [catalogName, setCatalogName] = useState("");
    let name = "";
    selectRows[0].cells.forEach((item)=>{
        if (item.id.split(":")[1] === "name"){
            name = item?.value;
        }
    });
    let navigate = useNavigate();

    const onSubmit = async () => {
        try {
          const {type, payload} = await deployCatalog({
            catalog_name:name,
            display_name:catalogName,
          }); // wait for the dispatch to complete
          if (type==="API_ERROR"){
            const errorTitle = "Catalog deployment failed"
            const errorMsg = payload.response.data.error;
            onError(errorTitle, errorMsg);
          }
        } catch (error) {
        console.log(error);
        }
        setActionProps("");
        navigate("/catalogs");
    };

  return (
    <Modal
      modalHeading="Deploy Catalog"
      danger={true}
      onRequestClose={()=>{
        setActionProps("");
      }}
      onRequestSubmit={()=>{
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
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
              onChange={(e)=>setCatalogName(e.target.value)}
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
