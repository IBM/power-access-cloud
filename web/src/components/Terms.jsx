import React from "react";
import { getTnCText, acceptTnC } from "../services/request";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { marked } from "marked";

import {
  Button,
  Grid,
  Column,
} from "@carbon/react";


const Terms = () => {
    const navigate = useNavigate();
    const [read, setRead] = useState("");

  useEffect(() => {
    const fetchText= async () => {
      let TnCText = await getTnCText();
 
      setRead(TnCText.text);
    }
    fetchText();
  }, []);

  return (

      <Grid fullWidth>
        <Column lg={16} md={8} sm={4} className="tnc" >
            
          <div
            dangerouslySetInnerHTML={{ __html: marked(read) }}
          />

          <Button onClick={async () => {
           await acceptTnC();
           navigate(-2);
          }}>I agree</Button>
        </Column>
      </Grid>
      
  );
};

export default Terms;
