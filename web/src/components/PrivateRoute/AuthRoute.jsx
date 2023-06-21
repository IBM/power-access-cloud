import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import UserService from "../../services/UserService";

const AuthRoute = ({ Component }) => {
  const navigate = useNavigate();
  const auth = UserService.isLoggedIn();
  
  useEffect(()=>{
    if (auth === false)
      navigate("/login");
  },[auth, navigate])

  return <Component />;
};

export default AuthRoute;
