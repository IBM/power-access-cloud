// import axios from "axios";
import React, { useState } from "react";
import { deployCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeployCatalog = ({selectRows,setActionProps})=> {
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
          await deployCatalog({
            catalog_name:name,
            display_name:catalogName,
          }); // wait for the dispatch to complete
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
              placeholder="Enter your name"
              name="name"
              value={catalogName}
              onChange={(e)=>setCatalogName(e.target.value)}
            />
          </div>
          <div className="mb-3">
            <label htmlFor="Name" className="form-label">
              Name
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
