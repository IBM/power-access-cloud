import React from "react";
import { Button } from "@carbon/react";
import UserService from "../services/UserService";
import { ArrowRight } from "@carbon/icons-react";

const ProfileSection = () => {
  return (
    <>
      <div className="user-profile">
        <div style={{ fontWeight: "bolder" }}>{UserService.getUsername()}</div>
        <div style={{ fontSize: "large" }}>{UserService.getName()}</div>
        <hr />
        <div>
          <Button
            kind="danger"
            size="lg"
            className="logout-button"
            style={{ width: "100%" }}
            onClick={() => UserService.doLogout()}
          >
            Log out <ArrowRight />
          </Button>
        </div>
      </div>
    </>
  );
};

export default ProfileSection;
