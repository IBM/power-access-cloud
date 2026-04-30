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
  TextInput,
  Checkbox,
  Modal,
  RadioButtonGroup,
  RadioButton
} from "@carbon/react";
import pacBackgroundImg from "../assets/images/PAC-background.jpg"

const Register = () => {
    const navigate = useNavigate();
    const [read, setRead] = useState("");
    const [open, setOpen] = useState(false);
    const [tnc_acc, setTnc_acc] = useState(false);
    
    // Form state
    const [formData, setFormData] = useState({
      userType: "",
      howDidYouHear: "",
      companyName: "",
      osProject: "",
      purposeReason: ""
    });

    // Track which fields have been touched (blurred)
    const [touchedFields, setTouchedFields] = useState({
      purposeReason: false
    });

  const handleInputChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleBlur = (field) => {
    setTouchedFields(prev => ({
      ...prev,
      [field]: true
    }));
  };

  const isFieldInvalid = (field, value) => {
    const MIN_CHARS = 100;
    const MAX_CHARS = 500;
    return touchedFields[field] && value.length > 0 && (value.length < MIN_CHARS || value.length > MAX_CHARS);
  };

  const getInvalidText = (field, value) => {
    const MIN_CHARS = 100;
    const MAX_CHARS = 500;
    if (value.length < MIN_CHARS) {
      return `Please provide at least ${MIN_CHARS} characters`;
    }
    if (value.length > MAX_CHARS) {
      return `Maximum ${MAX_CHARS} characters allowed`;
    }
    return "";
  };

  const handleUserTypeChange = (value) => {
    // Reset only conditional fields when user type changes
    // Keep purposeReason and howDidYouHear as they are common fields
    setFormData(prev => ({
      ...prev,
      userType: value,
      companyName: "",
      osProject: ""
    }));
  };

  const isFormValid = () => {
    if (!tnc_acc || !formData.userType) {
      return false;
    }

    // Character length validation for text areas
    const MIN_CHARS = 100;
    const MAX_CHARS = 500;
    const isPurposeValid = formData.purposeReason &&
                          formData.purposeReason.length >= MIN_CHARS &&
                          formData.purposeReason.length <= MAX_CHARS;

    // Validate based on user type
    switch(formData.userType) {
      case "isv":
        return formData.companyName && isPurposeValid;
      case "opensource":
        return formData.osProject && isPurposeValid;
      case "other":
        return isPurposeValid;
      default:
        return false;
    }
  };

  const generateJustificationText = () => {
    let justification = '';
    
    // Set user type label
    switch(formData.userType) {
      case "isv":
        justification = `User Type: ISV/IBM Customer/Partner\n`;
        break;
      case "opensource":
        justification = `User Type: OPEN SOURCE DEVELOPER\n`;
        break;
      case "other":
        justification = `User Type: OTHER\n`;
        break;
    }
    
    if (formData.howDidYouHear) {
      justification += `How did you hear about us: ${formData.howDidYouHear}\n`;
    }

    switch(formData.userType) {
      case "isv":
        justification += `Company Name: ${formData.companyName}\n`;
        break;
      case "opensource":
        justification += `Open Source Project/Community: ${formData.osProject}\n`;
        break;
      case "other":
        // No additional fields for "other" type
        break;
    }

    // Common field for all user types
    justification += `Purpose of using IBM® Power® Access Cloud: ${formData.purposeReason}`;

    return justification;
  };

  const handleClear = () => {
    setFormData({
      userType: "",
      howDidYouHear: "",
      companyName: "",
      osProject: "",
      purposeReason: ""
    });
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
              IBM&reg; Power&reg; Access Cloud
            </h1>
            <h1 className="landing-page__subheading banner-heading">
            Registration
            </h1>
          </Column>
          <Column lg={6} md={4} sm={4} >
            <img src={pacBackgroundImg} alt="ls" className="ls" />
          </Column>
       
        <Column lg={16} md={8} sm={4} className="tnc" >
       
<p>Please provide the following information for your access request. Provide as much detail as possible to help us process your request efficiently.
</p>
<p>
<strong>Note</strong>: By default, all new users are assigned to the Bronze group, which includes .5 vCPU, 8 GB of memory. <strong>IBM&reg; Power&reg; Access Cloud Groups</strong> control resource allocation by assigning the maximum CPU and memory available to your VM. With a valid use case, you can request additional resources from the IBM&reg; Power&reg; Access Cloud dashboard after your initial registration is approved.
</p>

{/* User Type Selection */}
<div style={{ marginBottom: '1.5rem' }}>
  <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
    I am a/an: <span className="text-danger">*</span>
  </label>
  <RadioButtonGroup
    name="userType"
    valueSelected={formData.userType}
    onChange={handleUserTypeChange}
    orientation="vertical"
  >
    <RadioButton labelText="ISV(Independent Software Vendor)/IBM Customer/Partner" value="isv" id="radio-isv" />
    <RadioButton labelText="Open Source Developer" value="opensource" id="radio-opensource" />
    <RadioButton labelText="Other" value="other" id="radio-other" />
  </RadioButtonGroup>
</div>

{/* Conditional Fields for ISV/IBM Customer/Partner */}
{formData.userType === "isv" && (
  <div style={{ marginBottom: '1.5rem' }}>
    <TextInput
      id="companyName"
      labelText={<>Company Name <span className="text-danger">*</span></>}
      placeholder="Enter your company name"
      value={formData.companyName}
      onChange={(e) => handleInputChange('companyName', e.target.value)}
    />
  </div>
)}

{/* Conditional Fields for Open Source Developer */}
{formData.userType === "opensource" && (
  <div style={{ marginBottom: '1.5rem' }}>
    <TextInput
      id="osProject"
      labelText={<>Open Source Project/Community <span className="text-danger">*</span></>}
      placeholder="Enter project or community name"
      value={formData.osProject}
      onChange={(e) => handleInputChange('osProject', e.target.value)}
    />
  </div>
)}

{/* Common Purpose Field - Always shown */}
<div style={{ marginBottom: '1.5rem' }}>
  <label htmlFor="purposeReason" className="cds--label">
    Purpose of using IBM&reg; Power&reg; Access Cloud <span className="text-danger">*</span>
  </label>
  <TextArea
    id="purposeReason"
    rows={6}
    placeholder="Please provide detailed information about your use case and how you plan to use IBM® Power® Access Cloud (e.g., product testing, porting, CI/CD, performance optimization, customer requirements)"
    value={formData.purposeReason}
    onChange={(e) => handleInputChange('purposeReason', e.target.value)}
    onBlur={() => handleBlur('purposeReason')}
    invalid={isFieldInvalid('purposeReason', formData.purposeReason)}
    invalidText={getInvalidText('purposeReason', formData.purposeReason)}
    helperText={`${formData.purposeReason.length}/500 characters (minimum 100 required)`}
    maxLength={500}
  />
</div>

{/* How did you hear about us - Optional, shown last */}
<div style={{ marginBottom: '1.5rem' }}>
  <TextInput
    id="howDidYouHear"
    labelText="How did you hear about us?"
    placeholder="e.g., IBM contact, conference, website, etc."
    value={formData.howDidYouHear}
    onChange={(e) => handleInputChange('howDidYouHear', e.target.value)}
  />
</div>

</Column>
       
        <Column lg={16} md={8} sm={4} >
          <p className="text">Read and accept the IBM&reg; Power&reg; Access Cloud usage terms and conditions and then click <strong>Submit</strong> to log into the IBM&reg; Power&reg; Access Cloud dashboard with your IBMid or GitHub account. You will be notified within 2 business days at the email you provide when your request is approved. You can also check status directly from the dashboard.
</p>
<span><span className="cb"><Checkbox labelText="" lg={1} md={1} sm={1} id="tnc_cb" disabled checked={tnc_acc} /></span><span>  I have read and accept the <a className="hyperlink" onClick={() => setOpen(true)} href="#/">IBM&reg; Power&reg; Access Cloud terms and conditions</a>.</span></span>
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
        <div className="last"><Button  kind="tertiary" onClick={handleClear}>Clear</Button>  <Button disabled={!isFormValid()} kind="primary" onClick={async () => {
            const justificationText = generateJustificationText();
            sessionStorage.setItem("Justification", justificationText);
           sessionStorage.setItem("TnC_acceptance", tnc_acc);
           navigate("/dashboard")
          }}>Submit</Button>
          </div>
        </Column>
      </Grid>
      
  );
};

export default Register;
