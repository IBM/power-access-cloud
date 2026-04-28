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
      ibmContact: "",
      companyName: "",
      architectureReason: "",
      osProject: "",
      vmReason: "",
      otherArchitectureReason: "",
      projectDescription: ""
    });

    // Track which fields have been touched (blurred)
    const [touchedFields, setTouchedFields] = useState({
      architectureReason: false,
      vmReason: false,
      projectDescription: false,
      otherArchitectureReason: false
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
    const MIN_CHARS = 25;
    return touchedFields[field] && value.length > 0 && value.length < MIN_CHARS;
  };

  const handleUserTypeChange = (value) => {
    // Reset conditional fields when user type changes
    setFormData(prev => ({
      ...prev,
      userType: value,
      ibmContact: "",
      companyName: "",
      architectureReason: "",
      osProject: "",
      vmReason: "",
      otherArchitectureReason: "",
      projectDescription: ""
    }));
  };

  const isFormValid = () => {
    if (!tnc_acc || !formData.userType) {
      return false;
    }

    // Minimum character length for text areas
    const MIN_CHARS = 25;

    // Validate based on user type
    switch(formData.userType) {
      case "isv":
        return formData.companyName &&
               formData.architectureReason &&
               formData.architectureReason.length >= MIN_CHARS;
      case "opensource":
        return formData.osProject &&
               formData.vmReason &&
               formData.vmReason.length >= MIN_CHARS;
      case "ibm":
        return formData.ibmContact &&
               formData.projectDescription &&
               formData.projectDescription.length >= MIN_CHARS;
      case "other":
        return formData.otherArchitectureReason &&
               formData.otherArchitectureReason.length >= MIN_CHARS;
      default:
        return false;
    }
  };

  const generateJustificationText = () => {
    let justification = `User Type: ${formData.userType.toUpperCase()}\n`;
    
    if (formData.howDidYouHear) {
      justification += `How did you hear about us: ${formData.howDidYouHear}\n`;
    }

    switch(formData.userType) {
      case "isv":
        justification += `Company Name: ${formData.companyName}\n`;
        justification += `Purpose of using Power VMs: ${formData.architectureReason}`;
        break;
      case "opensource":
        justification += `Open Source Project/Community: ${formData.osProject}\n`;
        justification += `Purpose of using Power VMs: ${formData.vmReason}`;
        break;
      case "ibm":
        if (formData.ibmContact) {
          justification += `IBM Contact/Sponsor: ${formData.ibmContact}\n`;
        }
        justification += `Project Description: ${formData.projectDescription}`;
        break;
      case "other":
        justification += `Purpose of using Power VMs: ${formData.otherArchitectureReason}`;
        break;
    }

    return justification;
  };

  const handleClear = () => {
    setFormData({
      userType: "",
      howDidYouHear: "",
      ibmContact: "",
      companyName: "",
      productName: "",
      architectureReason: "",
      osProject: "",
      githubProfile: "",
      vmReason: "",
      otherArchitectureReason: "",
      projectDescription: ""
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
    <RadioButton labelText="ISV (Independent Software Vendor)" value="isv" id="radio-isv" />
    <RadioButton labelText="Open Source Developer" value="opensource" id="radio-opensource" />
    <RadioButton labelText="IBM Customer/Partner" value="ibm" id="radio-ibm" />
    <RadioButton labelText="Other" value="other" id="radio-other" />
  </RadioButtonGroup>
</div>

{/* How did you hear about us - Optional for all */}
<div style={{ marginBottom: '1.5rem' }}>
  <TextInput
    id="howDidYouHear"
    labelText="How did you hear about us?"
    placeholder="e.g., IBM contact, conference, website, etc."
    value={formData.howDidYouHear}
    onChange={(e) => handleInputChange('howDidYouHear', e.target.value)}
  />
</div>

{/* Conditional Fields for ISV */}
{formData.userType === "isv" && (
  <>
    <div style={{ marginBottom: '1.5rem' }}>
      <TextInput
        id="companyName"
        labelText={<>Company Name <span className="text-danger">*</span></>}
        placeholder="Enter your company name"
        value={formData.companyName}
        onChange={(e) => handleInputChange('companyName', e.target.value)}
      />
    </div>
    <div style={{ marginBottom: '1.5rem' }}>
      <label htmlFor="architectureReason" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
        Purpose of using Power VM's <span className="text-danger">*</span>
      </label>
      <TextArea
        id="architectureReason"
        rows={4}
        placeholder="Tell us about your use case (e.g., product testing, customer requirements, performance optimization)"
        value={formData.architectureReason}
        onChange={(e) => handleInputChange('architectureReason', e.target.value)}
        onBlur={() => handleBlur('architectureReason')}
        invalid={isFieldInvalid('architectureReason', formData.architectureReason)}
        invalidText="Please provide at least 25 characters"
        helperText={`${formData.architectureReason.length} characters (minimum 25 required)`}
      />
    </div>
  </>
)}

{/* Conditional Fields for Open Source Developer */}
{formData.userType === "opensource" && (
  <>
    <div style={{ marginBottom: '1.5rem' }}>
      <TextInput
        id="osProject"
        labelText={<>Open Source Project/Community <span className="text-danger">*</span></>}
        placeholder="Enter project or community name"
        value={formData.osProject}
        onChange={(e) => handleInputChange('osProject', e.target.value)}
      />
      <p style={{ fontSize: '0.875rem', color: '#6f6f6f', marginTop: '0.25rem', fontStyle: 'italic' }}>
        Your GitHub profile will be linked from your login
      </p>
    </div>
    <div style={{ marginBottom: '1.5rem' }}>
      <label htmlFor="vmReason" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
        Purpose of using Power VM's <span className="text-danger">*</span>
      </label>
      <TextArea
        id="vmReason"
        rows={4}
        placeholder="Share how you plan to use Power VMs for your project (e.g., porting, testing, CI/CD)"
        value={formData.vmReason}
        onChange={(e) => handleInputChange('vmReason', e.target.value)}
        onBlur={() => handleBlur('vmReason')}
        invalid={isFieldInvalid('vmReason', formData.vmReason)}
        invalidText="Please provide at least 25 characters"
        helperText={`${formData.vmReason.length} characters (minimum 25 required)`}
      />
    </div>
  </>
)}

{/* Conditional Fields for IBM Customer/Partner */}
{formData.userType === "ibm" && (
  <>
    <div style={{ marginBottom: '1.5rem' }}>
      <TextInput
        id="ibmContact"
        labelText={<>IBM Contact/Sponsor (if any) <span className="text-danger">*</span></>}
        placeholder="Enter IBM contact name or email"
        value={formData.ibmContact}
        onChange={(e) => handleInputChange('ibmContact', e.target.value)}
      />
    </div>
    <div style={{ marginBottom: '1.5rem' }}>
      <label htmlFor="projectDescription" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
        Brief Project Description (1-2 sentences) <span className="text-danger">*</span>
      </label>
      <TextArea
        id="projectDescription"
        rows={4}
        placeholder="Provide a brief description of your project and how you plan to use IBM® Power® Access Cloud"
        value={formData.projectDescription}
        onChange={(e) => handleInputChange('projectDescription', e.target.value)}
        onBlur={() => handleBlur('projectDescription')}
        invalid={isFieldInvalid('projectDescription', formData.projectDescription)}
        invalidText="Please provide at least 25 characters"
        helperText={`${formData.projectDescription.length} characters (minimum 25 required)`}
      />
    </div>
  </>
)}

{/* Conditional Fields for Other */}
{formData.userType === "other" && (
  <div style={{ marginBottom: '1.5rem' }}>
    <label htmlFor="otherArchitectureReason" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
      Purpose of using Power VMs<span className="text-danger">*</span>
    </label>
    <TextArea
      id="otherArchitectureReason"
      rows={4}
      placeholder="Please explain purpose of access to Power Architecture resources"
      value={formData.otherArchitectureReason}
      onChange={(e) => handleInputChange('otherArchitectureReason', e.target.value)}
      onBlur={() => handleBlur('otherArchitectureReason')}
      invalid={isFieldInvalid('otherArchitectureReason', formData.otherArchitectureReason)}
      invalidText="Please provide at least 25 characters"
      helperText={`${formData.otherArchitectureReason.length} characters (minimum 25 required)`}
    />
  </div>
)}

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
