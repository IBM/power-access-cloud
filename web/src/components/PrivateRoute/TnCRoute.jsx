import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import UserService from "../../services/UserService";
import { tncStatus } from "../../services/request";

const TnCRoute = ({ Component }) => {
  const navigate = useNavigate();
  const auth = UserService.isLoggedIn();
  const [isLoading, setIsLoading] = useState(true);
    useEffect(()=>{
    const fetchStatus = async () => {
      let TnCdata = await tncStatus();
      
      if (!TnCdata.acceptance) {   
        navigate("/terms");
      }
      setIsLoading(false);
    };
    if (auth=== false){
      setIsLoading(false);
      navigate("/login");
    }else{
      fetchStatus();       
    }
  },[auth, navigate])


  if (isLoading) {
    return null; 
  }else{
    return <Component />;
  }
  
};

export default TnCRoute;
