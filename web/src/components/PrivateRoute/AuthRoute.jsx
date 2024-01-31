import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import UserService from "../../services/UserService";

const AuthRoute = ({ Component }) => {
  const navigate = useNavigate();
  const auth = UserService.isLoggedIn();
  useEffect(()=>{
    
      if (auth === false){

        const just = sessionStorage.getItem("Justification");
        const tnc_acc = sessionStorage.getItem("TnC_acceptance");
        if(just!=='' && tnc_acc){
          UserService.doLogin();
        }
      }
  },[auth, navigate])

    return <Component />;

};

export default AuthRoute;