import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import UserService from "../../services/UserService";
import axios from "axios";

const TnCRoute = ({ Component }) => {
  const navigate = useNavigate();
  const auth = UserService.isLoggedIn();
  const [isLoading, setIsLoading] = useState(true);
  const _axios = axios.create();
  _axios.interceptors.request.use((config) => {
    if (auth) {
      const cb = () => {
        config.headers.Authorization = `Bearer ${UserService.getToken()}`;
        return Promise.resolve(config);
      };
      return UserService.updateToken(cb);
    }
  });
  const url1 = "/pac-go-server/tnc";
   
   useEffect(()=>{
    if (auth=== false){
      setIsLoading(false);
      navigate("/login");
    }else{

       _axios.get(url1)
        .then((response) => {
          if (!response.data.accepted) {
            
            navigate("/terms");
          }
          setIsLoading(false);
        })
        .catch((error) => {
          console.log(error)
        });
        
    }
  },[auth, _axios, navigate])


  if (isLoading) {
    return null; 
  }else{
    return <Component />;
  }
  
};

export default TnCRoute;