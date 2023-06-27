import React, { useState } from "react";
import UserService from "../services/UserService";
import "../styles/welcome.scss";
import { ReactComponent as WelcomeImage } from "../assets/images/welcome.svg";

const Welcome = () => {
  const [termsAgreed, setTermsAgreed] = useState(false);

  const handleLogin = async () => {
    if (termsAgreed) {
      UserService.doLogin();
    } else {
      alert("Please agree to the terms and conditions.");
    }
  };

  const handleCheckboxChange = () => {
    setTermsAgreed(!termsAgreed);
  };

  return (
    <div className="jumbotron" style={{ backgroundColor: "#EAF6FA" }}>
      <div className="login-page">
        {/* WelcomeImage is downloaded from undraw.io */}
        {/* License: https://undraw.co/license */}
        <WelcomeImage className="welcome" alt="login" />
        <h1>Welcome to Power Access Cloud</h1>
        <div>
          <p>
            Please read and agree to the{" "}
            <a
              href="http://www.example.com"
              target="_blank"
              rel="noopener noreferrer"
            >
              terms and conditions
            </a>
            :
          </p>
          <label htmlFor="termsCheckbox" style={{ marginBottom: "10px" }}>
            <input
              type="checkbox"
              id="termsCheckbox"
              checked={termsAgreed}
              onChange={handleCheckboxChange}
            />
            <span style={{ marginLeft: "5px" }}>
              I agree to the terms and conditions
            </span>
          </label>
        </div>
        <button onClick={handleLogin}>Login</button>
      </div>
    </div>
  );
};

export default Welcome;
