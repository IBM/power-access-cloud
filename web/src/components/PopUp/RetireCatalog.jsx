// import axios from "axios";
import React from "react";
import { retireCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal } from "@carbon/react";

const RetireCatalog = ({ selectRows, setActionProps, response }) => {
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
      const { type, payload } = await retireCatalog(name); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "The retirement operation on the catalog has failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "The retirement operation on the catalog was successful.";
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
      modalHeading="Retire Catalog"
      danger={true}
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
    >
      <div>
        <div className="mb-3">
          <h4>Are you sure want to retire this Catalog?</h4>
        </div>
      </div>
    </Modal>
  );
};

export default RetireCatalog;
