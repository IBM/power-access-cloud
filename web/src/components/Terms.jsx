import React from "react";
import UserService from "../services/UserService";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { marked } from "marked";

import {
  Button,
  Grid,
  Column,
} from "@carbon/react";


const Terms = () => {
    const navigate = useNavigate();
  const url =
    "https://raw.githubusercontent.com/PDeXchange/pac-support/main/Terms%20and%20Conditions.md";
  const [read, setRead] = useState("");

  useEffect(() => {
    axios.get(url).then((response) => {
      setRead(response.data);
    });
  }, []);

  const _axios = axios.create();
  _axios.interceptors.request.use((config) => {
    if (UserService.isLoggedIn()) {
      const cb = () => {
        config.headers.Authorization = `Bearer ${UserService.getToken()}`;
        return Promise.resolve(config);
      };
      return UserService.updateToken(cb);
    }
  });
  const url1 = "/pac-go-server/tnc";

  return (

      <Grid fullWidth>
        <Column lg={16} md={8} sm={4} className="tnc" >
            
          <div
            dangerouslySetInnerHTML={{ __html: marked(read) }}
          />

          <Button onClick={async () => {
           await _axios.post(url1)
              .then((response) => {
                console.log(response.data)
              })
              .catch((error) => {
                console.log(error)
              });
           navigate(-2);
          }}>I agree</Button>
        </Column>
      </Grid>
      
  );
};

export default Terms;
