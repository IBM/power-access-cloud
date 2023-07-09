// import axios from "axios";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { newRequest, getGroup } from "../../services/request";
import { Modal } from "@carbon/react";

const NewRequest = ({ selectRows, setActionProps, response }) => {
  const [loading, setLoading] = useState(true);
  const id = selectRows[0]?.id;
  useEffect(() => {
    loadUser();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  let navigate = useNavigate();

  const [g, setGroup] = useState({
    name: "",
    id: "",
    justification: "",
  });

  const loadUser = async () => {
    const result = await getGroup(id);
    setLoading(false);
    setGroup({
      id: result?.payload?.id,
      name: result?.payload?.name,
      justification: "",
    });
  };

  const onInputChange = (e) => {
    setGroup({ ...g, [e.target.name]: e.target.value });
  };

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await newRequest(g); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "The request for the group has failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "The request for the group has been successfully submitted, track the progress under request section."
      }
    } catch (error) {
      // handle any errors that occurred during the dispatch
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate("/groups");
  };

  return (
    <Modal
      modalHeading="Request for a group"
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
      {loading && <>Loading....</>}
      {!loading && (
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
              value={g?.name}
              disabled
            />
          </div>
          <div className="mb-3">
            <label htmlFor="Justifcation" className="form-label">
              Justifcation<span className="text-danger">*</span>
            </label>
            <textarea
              type={"text"}
              className="form-control"
              placeholder="Enter your Justifcation for joining the group"
              name="justification"
              value={g?.justification}
              onChange={(e) => onInputChange(e)}
              required
            />
          </div>
        </div>
      )}
    </Modal>
  );
};

export default NewRequest;
