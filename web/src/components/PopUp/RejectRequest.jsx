// import axios from "axios";
import React, { useState } from "react";
import { rejectRequest } from "../../services/request";
import { Modal } from "@carbon/react";

const RejectRequest = ({ selectRows, setActionProps }) => {
  const [request, setRequest] = useState({
    id: selectRows[0]?.id,
    comment: "",
  });

  const onSubmit = async () => {
    try {
      await rejectRequest(request); // wait for the dispatch to complete
    } catch (error) {
      console.log(error);
    }
  };

  const onInputChange = (e) => {
    setRequest({ ...request, [e.target.name]: e.target.value });
  };
  return (
    <Modal
      modalHeading="Reject request"
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
          <h4>Are you sure want to Reject this request?</h4>
        </div>
        <div className="mb-3">
          <label htmlFor="Comment" className="form-label">
            Comment<span className="text-danger">*</span>
          </label>
          <textarea
            type={"text"}
            className="form-control"
            placeholder="Enter your comment here for rejecting the request"
            name="comment"
            value={request?.comment}
            onChange={(e) => onInputChange(e)}
            required
          />
        </div>
      </div>
    </Modal>
  );
};

export default RejectRequest;
