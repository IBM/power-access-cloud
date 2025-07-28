// import axios from "axios";
import React, { useState } from "react";
import { Modal } from "@carbon/react";
import { Theme } from "@carbon/react";
import { FaceNeutral, FaceDissatisfied, FaceSatisfied , FaceNeutralFilled, FaceDissatisfiedFilled, FaceSatisfiedFilled } from "@carbon/icons-react";
const Feedback = ({ setActionProps }) => {
  const [feedbackText,setFeedbackText]=useState("");
  const [feedbackRating,setFeedbackRating]=useState("Positive");

  const onSubmit = async () => {
    //alert(feedbackText + " "+ feedbackRating)
    window.location.href = `mailto:PowerACL@ibm.com?subject=IBM® Power® Access Cloud feedback&body=Rating:${feedbackRating}%0D%0AFeedback: ${feedbackText}`;
    setActionProps("")
     };
 
  return (
    <Theme theme="g10">
    <Modal
      modalHeading="Please provide your feedback"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        onSubmit();
      }}
      open={true}
      primaryButtonText={"Submit"}
      secondaryButtonText={"Cancel"}
    >
      <div>
      <div className="mb-3">
          <label htmlFor="rating" className="form-label">
            Rate your experience:
          </label>
          <div className="feedback">
<div onClick={() => setFeedbackRating("Negative") } style={{backgroundColor:"#F4BAAD"}} className={`${feedbackRating==="Negative" ? "ratingSelected" : "unselected"}`}>
    <div style={{marginBottom:"5px"}}>Negative</div> 
    {feedbackRating==="Negative"? <FaceDissatisfiedFilled size="50" style={{fill:"red"}}/>:<FaceDissatisfied style={{fill:"rgb(90, 90, 90)"}} size="50"/>}

</div>
<div  onClick={() => setFeedbackRating("Neutral") } style={{backgroundColor:"#FAFADF"}} className={`${feedbackRating==="Neutral" ? "ratingSelected" : "unselected"}`}>
    <div style={{marginBottom:"5px"}}>Neutral</div> 
    {feedbackRating==="Neutral"? <FaceNeutralFilled size="50" style={{fill:"#F4EE1F"}}/>:<FaceNeutral style={{fill:"rgb(90, 90, 90)"}} size="50"/>}
</div>
<div  onClick={() => setFeedbackRating("Positive") } style={{backgroundColor:"#CEF5CB"}} className={`${feedbackRating==="Positive" ? "ratingSelected" : "unselected"}`}>
<div style={{marginBottom:"5px"}}>Positive</div> 
{feedbackRating==="Positive"? <FaceSatisfiedFilled size="50" style={{fill:"green"}}/>:<FaceSatisfied style={{fill:"rgb(90, 90, 90)"}} size="50"/>}
</div>
          </div>
          </div>
        <div className="mb-3">
          <label htmlFor="comments" className="form-label">
            Comments (optional):
          </label>
          <textarea
            type={"text"}
            className="form-control"
            name="comments"
            rows="4"
            value= {feedbackText}
            onChange={(e) => {
                setFeedbackText(e.target.value)
            }}
          />
        </div>
      </div>
    </Modal>
    </Theme>
  );
}
export default Feedback;