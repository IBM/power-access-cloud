import React from "react";
import { getTnCText } from "../services/request";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { marked } from "marked";
import "../styles/registration.scss";
import {
  Button,
  Grid,
  Column,
  TextArea,
  Checkbox,
  Modal
} from "@carbon/react";

const Register = () => {
    const navigate = useNavigate();
    const [read, setRead] = useState("");
    const [open, setOpen] = useState(false);
    const [tnc_acc, setTnc_acc] = useState(false);
    const [just, setJust] = useState('');

  const handleChange = event => {
    setJust(event.target.value);

  };
    
  useEffect(() => {
    const fetchText= async () => {
      let TnCText = await getTnCText();
 
      setRead(TnCText.text);
    }
    fetchText();
  }, []);

  return (

      <Grid fullWidth >
       <Column lg={10} md={4} sm={4} className="info">
            <h1 className="landing-page__heading banner-heading">
              Power Access Cloud
            </h1>
            <h1 className="landing-page__subheading banner-heading">
            Registration
            </h1>
            
          </Column>
          <Column lg={6} md={4} sm={4} >
            <img src={`https://isv-graphics.s3.us-south.cloud-object-storage.appdomain.cloud/PAC-background-new.jpg`} alt="ls" className="ls" />
          </Column>
       
        <Column lg={16} md={8} sm={4} className="tnc" >
       
<p>Briefly describe how you plan to use this service. Provide as much detail as possible and include a link to your project if you have one.
</p>
<p>
<strong>Note</strong>: By default, all new users are assigned to the Bronze group, which includes .5 vCPU, 8 GB of memory. <strong>PAC Groups</strong> control resource allocation by assigning the maximum CPU and memory available to your VM. With a valid use case, you can request additional resources from the PAC dashboard after your initial registration is approved.

</p>
       
<TextArea labelText="Tell us how you plan to use your PAC service" rows={4} id="justification" onChange={handleChange} value={just} />
</Column>
       
        <Column lg={16} md={8} sm={4} >
          <p className="text">Read and accept the Power Access Cloud usage terms and conditions and then click <strong>Submit</strong> to log into the PAC dashboard with your IBMid or GitHub account. You will be notified within 2 business days at the email you provide when your request is approved. You can also check status directly from the dashboard.
</p>
<span><span className="cb"><Checkbox labelText="" lg={1} md={1} sm={1} id="tnc_cb" disabled checked={tnc_acc} /></span><span>  I have read and accept the <a className="hyperlink" onClick={() => setOpen(true)}>Power Access Cloud terms and conditions</a>.</span></span>
<Modal 
aria-label=""
size="lg" 
      open={open} 
      onRequestClose={() => {setTnc_acc(false); setOpen(false)}} 
      hasScrollingContent 
      primaryButtonText={"Accept"}
      secondaryButtonText={"Cancel"}
      onRequestSubmit={() => {
        setTnc_acc(true); setOpen(false); 
      }}>
<div className="tnc" dangerouslySetInnerHTML={{ __html: marked(read) }}></div>

</Modal>
          
<br/><br/>
        <div className="last"><Button  kind="tertiary" onClick={() => {
           setJust('');
          }}>Clear</Button>  <Button disabled={!(tnc_acc && (just!==''))} kind="primary" onClick={async () => {
            sessionStorage.setItem("Justification", just);
           sessionStorage.setItem("TnC_acceptance", tnc_acc);
           navigate("/dashboard")
          }}>Submit</Button>
          </div>
        </Column>
      </Grid>
      
  );
};

export default Register;
