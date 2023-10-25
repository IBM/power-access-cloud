import React from "react";
import UserService from "../services/UserService";
import "../styles/welcome.scss";
import { ReactComponent as WelcomeImage } from "../assets/images/welcome.svg";

const Welcome = () => {
 
  const handleLogin = async () => {
    
      UserService.doLogin();
   
  };

  return (
    <div className="jumbotron" style={{ backgroundColor: "#EAF6FA" }}>
      <div className="login-page">
        {/* WelcomeImage is downloaded from undraw.io */}
        {/* License: https://undraw.co/license */}
        <WelcomeImage className="welcome" alt="login" />
        <h1>Welcome to Power Access Cloud</h1>
        
        <button onClick={handleLogin}>Login</button>
      </div>
    </div>
  );
};

export default Welcome;
