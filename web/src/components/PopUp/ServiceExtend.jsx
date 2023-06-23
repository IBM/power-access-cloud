// import axios from "axios";
import React, {useState} from "react";
import { extendServices } from "../../services/request";
import { useNavigate } from "react-router-dom";
import { Modal,DatePicker,DatePickerInput } from "@carbon/react";

const ServiceExtend = ({selectRows,setActionProps})=> {
    const [date, setDate] = useState(null);
    const [justification, setJustification] = useState("");
    let name = "";
    selectRows[0].cells.forEach((item)=>{
        if (item.id.split(":")[1] === "name"){
            name = item?.value;
        }
    });
  let navigate = useNavigate();

  const onSubmit = async () => {
    const changedDate = new Date(date);
    const isoString = changedDate.toISOString();
    try {
      await extendServices(name, {
        justification,
        type:"SERVICE_EXPIRY",
        service:{
            expiry: isoString
        }
      }); // wait for the dispatch to complete
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
      onRequestClose={()=>{
        setActionProps("");
      }}
      onRequestSubmit={()=>{
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
    >
      <div>
            <div className="mb-3">
                <label htmlFor="Name" className="form-label">
                    Comment
                </label>
                <input
                type={"text"}
                className="form-control"
                placeholder="Comment for extend"
                name="justification"
                value={justification}
                onChange={(e)=> setJustification(e.target.value)}
                />
            </div>
            <DatePicker 
                locale="no" 
                dateFormat="d/m/Y" 
                datePickerType="single"
                onChange={(value)=>{
                    console.log(value);
                    setDate(value);
                }}
            >
                <DatePickerInput placeholder="dd/mm/yyyy" />
            </DatePicker>
      </div>
    </Modal>
  );
}

export default ServiceExtend;
