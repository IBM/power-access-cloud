// import axios from "axios";
import React from "react";
import { Modal, Theme } from "@carbon/react";
import { deleteUser } from "../../services/request";
import UserService from "../../services/UserService";

const DeleteAccount = ({ setActionProps }) => {

	const onSubmit = async () => {
	let title = "";
	let message = "";
	let errored = false;
	const userName = UserService.getUsername()

		try {
				const { type, payload } = await deleteUser(userName); // wait for the dispatch to complete
				if (type === "API_ERROR") {
				title = "Service deletion failed.";
				message = payload.response.data.error;
				errored = true;
				} else {
				title = "Service deleted successfully.";
				UserService.doLogout()
				}
		} catch (error) {
				console.log("Delete failed : ", error)
		}
		setActionProps("")
	};
 
  return (
    <Theme theme="g10">
    <Modal
			modalLabel="Delete account"
      modalHeading="Are you sure you want to delete your account?"
      onRequestClose={() => {
        setActionProps("");
      }}
      onRequestSubmit={() => {
        onSubmit();
      }}
      open={true}
			danger={true}
      primaryButtonText={"Delete"}
      secondaryButtonText={"Cancel"}
    >
			<div>
        <div className="mb-3">
          <h6>This will delete all your existing services, keys and revoke access from your group.</h6>
        </div>
      </div>
    </Modal>
    </Theme>
  );
}
export default DeleteAccount;