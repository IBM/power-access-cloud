// import axios from "axios";
import React, { useState } from "react";
import { extendServices } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal, DatePicker, DatePickerInput } from "@carbon/react";

const ServiceExtend = ({ selectRows, setActionProps, onError }) => {
  const [justification, setJustification] = useState("");
  let name = "";
  let expiry = "";
  selectRows[0].cells.forEach((item) => {
    if (item.id.split(":")[1] === "name") {
      name = item?.value;
    } else if (item.id.split(":")[1] === "expiry") {
      expiry = new Date(item?.value);
    }
  });
  const [date, setDate] = useState(expiry);

  let navigate = useNavigate();

  const onSubmit = async () => {
    const changedDate = new Date(date);
    const isoString = changedDate.toISOString();
    try {
      const {type, payload} = await extendServices(name, {
        justification,
        type: "SERVICE_EXPIRY",
        service: {
          expiry: isoString,
        },
      }); // wait for the dispatch to complete
      if (type==="API_ERROR"){
        const errorTitle = "Service extension failed"
        const errorMsg = payload.response.data.error;
        onError(errorTitle, errorMsg);
      }
    } catch (error) {
      console.log(error);
    }
    setActionProps("");
    navigate("/services");
  };

  return (
    <Modal
      modalHeading="Extend the service"
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
          <label htmlFor="Name" className="form-label">
            Justification<span className="text-danger">*</span>
          </label>
          <input
            type={"text"}
            className="form-control"
            placeholder="Justification to extend the expiry for the sevice"
            name="justification"
            value={justification}
            onChange={(e) => setJustification(e.target.value)}
          />
        </div>
        <DatePicker
          allowInput={true}
          locale="en"
          dateFormat="m/d/Y"
          value={date}
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
