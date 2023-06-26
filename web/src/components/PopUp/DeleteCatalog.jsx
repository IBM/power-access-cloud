// import axios from "axios";
import React from "react";
import { deleteCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteCatalog = ({selectRows,setActionProps})=> {
    let name = "";
    selectRows[0].cells.forEach((item)=>{
        if (item.id.split(":")[1] === "name"){
            name = item?.value;
        }
    });
  let navigate = useNavigate();

  const onSubmit = async () => {
    try {
      await deleteCatalog(name); // wait for the dispatch to complete
    } catch (error) {
      console.log(error);
    }
    setActionProps("");
    navigate("/catalogs");
  };

  return (
    <Modal
      modalHeading="Delete Catalog"
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
            <h4>Are you sure want to delete this Catalog!</h4>
          </div>
      </div>
    </Modal>
  );
}

export default DeleteCatalog;
