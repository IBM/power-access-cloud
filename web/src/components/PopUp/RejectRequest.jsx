// import axios from "axios";
import React, { useState } from "react";
import { rejectRequest } from "../../services/request";
import { Modal } from "@carbon/react";
import { useNavigate } from "react-router-dom";

const RejectRequest = ({ selectRows, setActionProps, response }) => {
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Reject");
  const [request, setRequest] = useState({
    id: selectRows[0]?.id,
    comment: "",
  });

  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await rejectRequest(request); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Rejection of request failed";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Request rejected successfully.";
      }
    } catch (error) {
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/requests");
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
        setPrimaryButtonDisabled(true);
        setPrimaryButtonText("Rejecting...")
        onSubmit();
      }}
      open={true}
      primaryButtonText={primaryButtonText}
      secondaryButtonText={"Cancel"}
      primaryButtonDisabled={primaryButtonDisabled}
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
