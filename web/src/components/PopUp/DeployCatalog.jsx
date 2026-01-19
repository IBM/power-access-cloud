import React, { useState } from "react";
import { deployCatalog } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal,InlineNotification } from "@carbon/react";

const MAX_NAME_LENGTH = 125;

const DeployCatalog = ({ selectRows, setActionProps, response }) => {
  const [catalogName, setCatalogName] = useState("");
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");
  const [emptyServiceName, setEmptyServiceName] = useState(true)
  const getCellValue = (data, key) => {
    if (!data) return '';
    
    // Admin Persona (Carbon Table format)
    if (data.cells) {
      const cell = data.cells.find(c => c.id.endsWith(`:${key}`));
      return cell ? cell.value : '';
    }

    // User Persona (Standard Object format)
    return data[key] || '';
  };
  const name = getCellValue(selectRows, 'name');
  let navigate = useNavigate();
  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    try {
      const { type, payload } = await deployCatalog({
        catalog_name: name,
        display_name: catalogName,
      }); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "The deployment of the catalog has failed.";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "The catalog was deployed successfully.";
        
      }
    } catch (error) {
      console.log(error);
    }
    response(title, message, errored)
    
    setActionProps("");

    if(errored){
      navigate("/catalogs");
    }else{
      setTimeout(() => {
        navigate("/");
      }, 2000);
      
    }
    
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
      modalHeading="Deploy Catalog"
      
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        if(catalogName===""){
          setEmptyServiceName(false)
          return;
        }
        if (catalogName.length > MAX_NAME_LENGTH){
          return;
        }
          setEmptyServiceName(true);
          setPrimaryButtonDisabled(true);
        setPrimaryButtonText("Submitting...")
        onSubmit();
        
        
      }}
      open={true}
      primaryButtonText={primaryButtonText}
      secondaryButtonText={"Cancel"}
      primaryButtonDisabled={primaryButtonDisabled}
    >
      <div>
        <div className="mb-3">
          <label htmlFor="Name" className="form-label">
            Name <span className="text-danger">*</span>
          </label>
          <input
            type={"text"}
            className="form-control"
            placeholder="Enter the display name for the service"
            name="name"
            value={catalogName}
            maxLength={MAX_NAME_LENGTH}
            onChange={(e) => {
              if(e.target.value===""){
                setEmptyServiceName(false)
              }else{
                setEmptyServiceName(true)

              }
              setCatalogName(e.target.value)
            }}
          />
          {catalogName.length}/{MAX_NAME_LENGTH} characters
          {!emptyServiceName&&inlineNotificationComponent('Service name',': field can not be empty')}

        </div>
        <div className="mb-3">
          <label htmlFor="Name" className="form-label">
            Catalog
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