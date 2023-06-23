// import axios from "axios";
import React from "react";
import { deleteKeys } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const DeleteKey = ({selectRows,setActionProps})=> {
  const id = selectRows[0]?.id;
  let navigate = useNavigate();

  const onSubmit = async () => {
    try {
      await deleteKeys({ id }); // wait for the dispatch to complete
    } catch (error) {
      console.log(error);
    }
    setActionProps("");
    navigate("/keys");
  };

  return (
    <Modal
      modalHeading="Exit request"
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
            <h4>Are you sure want to delete this key!</h4>
          </div>
      </div>
    </Modal>
  );
}

export default DeleteKey;
