// import axios from "axios";
import React, { useState } from "react";
import { deleteGroup } from "../../services/request";
import { Modal,InlineNotification } from "@carbon/react";
import { useNavigate } from "react-router-dom";

const ExitGroup = ({ selectRows, pagename, setActionProps, response }) => {
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");
  const [reasonCheck, setReasonCheck] = useState(false)
  const id = selectRows[0]?.id;
  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    console.log()
    if(g.justification===""){
      setReasonCheck(true)
      return;
    }
    
      try {
        const { type, payload } = await deleteGroup(g); // wait for the dispatch to complete
        if (type === "API_ERROR") {
          title = "The request to exit the group has failed.";
          message = payload.response.data.error;
          errored = true;
        } else {
          title = "Your request to leave the group was successfully submitted.";
        }
      } catch (error) {
        console.log(error);
      }
      response(title, message, errored)
      setActionProps("");
      setPrimaryButtonDisabled(true);
      setPrimaryButtonText("Exiting...")
      navigate(pagename);
    
  };

  const [g, setGroup] = useState({
    id: id,
    justification: "",
  });

  const onInputChange = (e) => {
    setGroup({ ...g, [e.target.name]: e.target.value });
  };
  const inlineNotificationComponent = (notificationTitile,notificationSubTitile) => {
    return (
      <React.Fragment>
        <InlineNotification
        aria-label="closes notification"
        kind="error"
        onClose={function noRefCheck(){}}
        onCloseButtonClick={function noRefCheck(){}}
        statusIconDescription="notification"
        subtitle={notificationSubTitile}
        title={notificationTitile}
      />
      </React.Fragment>
    );
  };
  return (
    <Modal
      modalHeading="Leave Group"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        
        onSubmit();
        
      }}
      open={true}
      primaryButtonText={primaryButtonText}
      primaryButtonDisabled={primaryButtonDisabled}
      secondaryButtonText={"Cancel"}
    >
      <div>
       
        <div className="mb-3">
          <label htmlFor="Justifcation" className="form-label">
          Please share why you are leaving this group<span className="text-danger">*</span>
          </label>
          <textarea
            placeholder="Enter your justifcation to exit from this group"
            type={"text"}
            className="form-control"
            name="justification"
            value={g?.justification}
            onChange={(e) => {
              if(e.target.value===''){
                setReasonCheck(true)
              }else{
                setReasonCheck(false)
              }
              onInputChange(e)
            }}
            onBlur={(e)=>{
              if(e.target.value===''){
                setReasonCheck(true)
              }else{
                setReasonCheck(false)
              }
            }}
            
            required
          />
          {reasonCheck && inlineNotificationComponent('Reason',': field can not be empty')}
        </div>
      </div>
    </Modal>
  );
};

export default ExitGroup;
