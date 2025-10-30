import React, { useState } from "react";
import { Link } from "@carbon/icons-react";
import Feedback from "./PopUp/Feedback";
import NewUserGuide from "./PopUp/NewUserGuide";
import ToastNotify from "./utils/ToastNotify";

const BUTTON_FEEDBACK = "BUTTON_FEEDBACK";
const BUTTON_USER_GUIDE = "BUTTON_USER_GUIDE";

const QuickLinks = () => {
  const [actionProps, setActionProps] = useState("");
  // notify
  const [notifyKind, setNotifyKind] = useState("");
  const [title, setTitle] = useState("");
  const [message, setMessage] = useState("");

  const action = {
    key: BUTTON_FEEDBACK,
    label: "Feedback",
  };

  const userGuideAction = {
    key: BUTTON_USER_GUIDE,
    label: "User Guide",
  };

  const handleFeedBackResponse = (title, message, errored) => {
    setTitle(title);
    setMessage(message);
    errored ? setNotifyKind("error") : setNotifyKind("success");
  };

  const renderActionModals = () => {
    return (
      <React.Fragment>
        {actionProps?.key === BUTTON_FEEDBACK && (
          <Feedback
            setActionProps={setActionProps}
            response={handleFeedBackResponse}
          />
        )}
        {actionProps?.key === BUTTON_USER_GUIDE && (
          <NewUserGuide setActionProps={setActionProps} />
        )}
      </React.Fragment>
    );
  };
  return (
    <>
      {renderActionModals()}
      <div
        style={{
          padding: "1rem",
          border: "1px solid #E4E5E6",
          minHeight: "22rem",
          overflow: "hidden",
        }}
      >
        <h4>Quick links</h4>
        <div style={{ marginTop: "2rem" }}>
          <Link size="20" />{" "}
          <a href="https://community.ibm.com/community/user/powerdeveloper/home">
            Join the Power Developer eXchange Community
          </a>
          <br />
          <br />
          <Link size="20" />{" "}
          <a
            href="https://github.com/PDeXchange/pac-support/blob/main/docs/FAQ.md"
            target="_blank"
            rel="noopener noreferrer"
          >
            Read the FAQ
          </a>
          <br />
          <br />
          <Link size="20" />{" "}
          <a
            style={{
              color: "rgb(13, 110, 253)",
              textDecoration: "underline",
              cursor: "pointer",
            }}
            onClick={() => setActionProps(action)}
            href="#/"
          >
            Provide feedback
          </a>
          <ToastNotify
            title={title}
            subtitle={message}
            kind={notifyKind}
            setTitle={setTitle}
          />
          <br />
          <br />
          <Link size="20" />{" "}
          <a
            style={{
              color: "rgb(13, 110, 253)",
              textDecoration: "underline",
              cursor: "pointer",
            }}
            onClick={() => setActionProps(userGuideAction)}
            href="#/"
          >
            User Guide
          </a>
        </div>
      </div>
    </>
  );
};

export default QuickLinks;
