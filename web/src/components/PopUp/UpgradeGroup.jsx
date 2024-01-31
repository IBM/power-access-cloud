// import axios from "axios";
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { newRequest, getGroup } from "../../services/request";
import { Modal,Select,SelectItem } from "@carbon/react";

const UpgradeGroup = ({pagename, currentGroupId, allgroupdata, setActionProps, response }) => {
  const [loading, setLoading] = useState(true);
  const id = currentGroupId;
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");

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
        title = "The request for the group has been successfully submitted, track the progress in My group section."
      }
    } catch (error) {
      // handle any errors that occurred during the dispatch
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate(pagename);
  };

  return (
    <Modal
      modalHeading="Upgrade group"
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
      {loading && <>Loading....</>}
      {!loading && (
        <div>
          <div className="mb-3">
            {/* <label htmlFor="Name" className="form-label">
              Name
            </label> */}
            <label htmlFor="Justifcation" className="form-label">
            Name<span className="text-danger">*</span>
            </label>
            {/* <input
              type={"text"}
              className="form-control"
              placeholder="Enter your name"
              name="name"
              value={g?.name}
              disabled
            /> */}
            <Select id="select-1" onChange={(e)=>{
                // alert(e.target.value)
                setGroup({ ...g, "id": e.target.value });
            }} labelText="Select a desired group" >
               
                {               
                allgroupdata.map((group, index) => (
                    <SelectItem key={group.id} value={group.id} text={`${group.name} (${group.quota.cpu} vCPU, ${group.quota.memory} GB)`} />
                ))} 
                
            </Select>
          </div>
          <div className="mb-3">
            <label htmlFor="Justifcation" className="form-label">
              Justification<span className="text-danger">*</span>
            </label>
            <textarea
              type={"text"}
              className="form-control"
              placeholder="Enter your justification for upgrading to this group"
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

export default UpgradeGroup;
