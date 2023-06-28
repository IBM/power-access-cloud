// import axios from "axios";
import React, { useState } from "react";
import { deleteGroup } from "../../services/request";
import { Modal } from "@carbon/react";

const ExitGroup = ({ selectRows, setActionProps }) => {
  const id = selectRows[0]?.id;

  const onSubmit = async () => {
    try {
      await deleteGroup(g); // wait for the dispatch to complete
    } catch (error) {
      console.log(error);
    }
  };

  const [g, setGroup] = useState({
    id: id,
    justification: "",
  });

  const onInputChange = (e) => {
    setGroup({ ...g, [e.target.name]: e.target.value });
  };

  return (
    <Modal
      modalHeading="Exit Group"
      danger={true}
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        onSubmit();
        setActionProps("");
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
    >
      <div>
        <div className="mb-3">
          <h4>Are you sure want to Exit from this group!</h4>
        </div>
        <div className="mb-3">
          <label htmlFor="Justifcation" className="form-label">
            Justifcation<span className="text-danger">*</span>
          </label>
          <textarea
            type={"text"}
            className="form-control"
            placeholder="Enter your justifcation to exit from this group"
            name="justification"
            value={g?.justification}
            onChange={(e) => onInputChange(e)}
            required
          />
        </div>
      </div>
    </Modal>
  );
};

export default ExitGroup;
