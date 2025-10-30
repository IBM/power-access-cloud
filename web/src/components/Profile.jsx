import React, { useState} from 'react'
import { Button } from "@carbon/react";
import UserService from "../services/UserService";
import { ArrowRight } from "@carbon/icons-react";
import DeleteAccount from './PopUp/DeleteAccount';

const BUTTON_DELETE_ACCOUNT = "BUTTON_DELETE_ACCOUNT"

const ProfileSection = () => {
  const [actionProps, setActionProps] = useState("");

  const deleteAccountAction={
    key: BUTTON_DELETE_ACCOUNT,
      label: "Delete Account"
  };

  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_DELETE_ACCOUNT && (
          <DeleteAccount
            setActionProps={setActionProps}
          />
        )}
        </React.Fragment>)}

  return (
    <>
      {renderActionModals()}
      <div className="user-profile">
        <div style={{ fontWeight: "bolder" }}>{UserService.getUsername()}</div>
        <div style={{ fontSize: "large" }}>{UserService.getName()}</div>
        <hr />
        <div>
          <Button
            kind="secondary"
            size="lg"
            className="delete-button"
            style={{ width: "100%" }}
            onClick={() => setActionProps(deleteAccountAction)}
            >
            Delete Account <ArrowRight />
          </Button>
        </div>   
        <div>
          <Button
            kind="danger"
            size="lg"
            className="logout-button"
            style={{ width: "100%" }}
            onClick={() => { 
              sessionStorage.setItem("Justification", '');
              sessionStorage.setItem("TnC_acceptance", false);
              UserService.doLogout()
            }}
          >
            Log out <ArrowRight />
          </Button>
        </div>
      </div>
    </>
  );
};

export default ProfileSection;
