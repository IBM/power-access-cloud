import React, { useState } from "react";
import { Modal, Theme } from "@carbon/react";
import { deleteUser } from "../../services/request";

const DeleteUser = ({ user, setActionProps, response, fetchData }) => {
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    if (submitting) {
      return;
    }

    setSubmitting(true);
    let title = "";
    let message = "";
    let errored = false;

    try {
      const { type, payload } = await deleteUser(user.id);
      if (type === "API_ERROR") {
        title = "User deletion failed.";
        message = payload.response?.data?.error || "An error occurred";
        errored = true;
      } else {
        title = "User deleted successfully.";
        message = `User ${user.username} has been deleted.`;
        await fetchData(); // Refresh the user list
      }
    } catch (error) {
      title = "User deletion failed.";
      message = error.message || "An unexpected error occurred";
      errored = true;
    }

    setSubmitting(false);
    response(title, message, errored);
    setActionProps("");
  };

  return (
    <Theme theme="g10">
      <Modal
        modalLabel="Delete User"
        modalHeading={`Are you sure you want to delete user "${user.username}"?`}
        onRequestClose={() => {
          if (!submitting) {
            setActionProps("");
          }
        }}
        onRequestSubmit={onSubmit}
        open={true}
        danger={true}
        primaryButtonText={submitting ? "Deleting..." : "Delete"}
        secondaryButtonText={"Cancel"}
        primaryButtonDisabled={submitting}
      >
        <div>
          <div className="mb-3">
            <p><strong>User ID:</strong> {user.id}</p>
            <p><strong>Name:</strong> {user.firstname} {user.lastname}</p>
            <p><strong>Email:</strong> {user.email}</p>
            <br />
            <p>This action will delete the user and revoke all their access.</p>
          </div>
        </div>
      </Modal>
    </Theme>
  );
};

export default DeleteUser;
