import React from "react";
import {
  Button,
} from "react-bootstrap";
import UserService from "../services/UserService";

const ProfileSection = ()=>{
  return (
    <div className="user-profile">
        <p>Username: {UserService.getName()}{UserService.getUsername()} </p>
        <Button variant="danger" style={{ marginLeft: "10px" }} onClick={() => UserService.doLogout()}>Logout</Button>
    </div>
  );
}

export default ProfileSection;