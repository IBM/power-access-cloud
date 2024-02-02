// import axios from "axios";
import React, { useState } from "react";
import { extendServices } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal, DatePicker, DatePickerInput } from "@carbon/react";

const ServiceExtend = ({ pagename, selectRows, setActionProps, response }) => {
  console.log(selectRows.display_name)
  const [primaryButtonDisabled, setPrimaryButtonDisabled] = useState(false);
  const [primaryButtonText, setPrimaryButtonText] = useState("Submit");
  const name = selectRows[0]?.name;
  const [justification, setJustification] = useState("");
  let expiry = "";
  // selectRows[0].cells.forEach((item) => {
  //   if (item.id.split(":")[1] === "expiry") {
  //     expiry = new Date(item?.value);
  //   }
  // });

  expiry=new Date(selectRows[0].expiry.split(" ")[0])

  console.log(expiry)
  const [date, setDate] = useState(expiry);

  let navigate = useNavigate();

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;
    const changedDate = new Date(date);
    const isoString = changedDate.toISOString();
    try {
      const { type, payload } = await extendServices(name, {
        justification,
        type: "SERVICE_EXPIRY",
        service: {
          expiry: isoString,
        },
      }); // wait for the dispatch to complete
      if (type === "API_ERROR") {
        title = "Service extension failed";
        message = payload.response.data.error;
        errored = true;
      } else {
        title = "Service extention request submitted successfully, please wait for the approval. For more details please check the Requests tab.";
      }
    } catch (error) {
      console.log(error);
    }
    response(title, message, errored)
    setActionProps("");
    navigate(pagename);
  };

  return (
    <Modal
      modalHeading="Extend the service"
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
      <div>
        <div className="mb-3">
          <label htmlFor="Name" className="form-label">
            Justification<span className="text-danger">*</span>
          </label>
          <input
            type={"text"}
            className="form-control"
            placeholder="Enter your justification for extending the service."
            name="justification"
            value={justification}
            onChange={(e) => setJustification(e.target.value)}
          />
        </div>
        <label htmlFor="Name" className="form-label">
            Select date<span className="text-danger">*</span>
          </label>
        <DatePicker
          allowInput={true}
          locale="en"
          dateFormat="m/d/Y"
          value={date}
          minDate={new Date()}
          datePickerType="single"
          onChange={(value) => {
            setDate(value);
          }}
        >
          <DatePickerInput placeholder="dd/mm/yyyy" />
        </DatePicker>
      </div>
    </Modal>
  );
};

export default ServiceExtend;
