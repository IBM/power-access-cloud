import React from "react";
import UserService from "../services/UserService";
import { useNavigate } from "react-router-dom";
import {
  Button,
  Grid,
  Column,
} from "@carbon/react";
import "../styles/welcome.scss";
import pacBackgroundImg from "../assets/images/PAC-background.jpg"

const Welcome = () => {
  let navigate = useNavigate();
  const handleLogin = async () => {
    
      UserService.doLogin()
      
   
  }
  const handleReg = async () => {
    
    navigate("/register");
 
};

  return (
    <Grid className="login-page" fullWidth>
      <Column lg={10} md={4} >
      <img src={pacBackgroundImg} alt="graphic" className="graphic" />
      </Column>
      <Column lg={6} md={4} sm={4} >
        <h3 className="landing-page__sub_heading">
        Thank you for your interest in IBM&reg; Power&reg; Access Cloud

        </h3>
   <br/>
        <p>IBM&reg; Power&reg; Access Cloud offers a comprehensive catalog of free Power hardware resources for developers who want to explore, develop, and optimize their open source software on the Power architecture. 
</p>
        <p>Our self-service dashboard allows you to select the resources you need and deploy them according to your specific requirements. After you register, you'll have access to your resources for 5 days with an option to extend if needed. Read our <a href="https://github.com/PDeXchange/pac-support/blob/main/docs/FAQ.md" target="_blank" rel="noopener noreferrer">FAQ</a> to learn more.
</p>
        <p>Click <strong>Register</strong> to get started. You'll need a valid <a href="https://www.ibm.com/account/us-en/signup/register.html" target="_blank" rel="noopener noreferrer">IBMid</a> or <a href="https://github.com/join" target="_blank" rel="noopener noreferrer">GitHub</a> account to complete the registration. If you've already registered, click <strong>Log in</strong> to go directly to the IBM&reg; Power&reg; Access Cloud dashboard. 
</p>
<br/>
<Button kind="tertiary" onClick={handleReg}>Register</Button> <Button kind="primary" onClick={handleLogin}>Log in</Button>
      </Column>
      </Grid>
    
  );
};

export default Welcome;
