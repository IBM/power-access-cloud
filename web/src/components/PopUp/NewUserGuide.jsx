// import axios from "axios";
import React from "react";
import { Modal, Button } from "@carbon/react";


const NewUserGuide = ({ setActionProps }) => {

  return (
    
    <Modal
      modalHeading="User Guide"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        alert('hi')
      }}
      open={true}
      primaryButtonText={"Ok"}
      secondaryButtonText={"Cancel"}
    >
      <div>
      <div className="mb-3">
          <h6>Step: 1</h6>
          Lorem ipsum dolor, sit amet consectetur adipisicing elit. Optio architecto illo quos, quidem delectus similique fuga ratione laboriosam pariatur? Natus earum architecto, nobis impedit nemo doloremque id officiis corporis reiciendis?
          <br />
          <h6>Step 2</h6>
          Lorem ipsum dolor sit amet consectetur, adipisicing elit. Saepe officiis odit quae hic ab, laudantium accusamus? Iure iusto saepe aut, accusantium sed non alias natus doloribus quod veniam ex tempore.
           <br />
          <h6>Step 3: </h6>
          Lorem ipsum dolor sit amet consectetur, adipisicing elit. Mollitia exercitationem et voluptatum asperiores non maxime, aperiam rem amet optio soluta, tenetur corporis vero ratione eum laborum suscipit saepe voluptatibus deserunt?
      </div>
      </div>
    </Modal>
    
  );
}
export default NewUserGuide;