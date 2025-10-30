import React, { useState, useEffect } from "react";
import { Modal } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import { getQuota } from "../../services/request";

const QuotaWarning = () => {
  const [showModal, setShowModal] = useState(false);
  const [quotaMessage, setQuotaMessage] = useState("");
  const [primaryButtonText, setPrimaryButtonText] = useState("");
  const [navigateTo, setNavigateTo] = useState("");
  let navigate = useNavigate();

  useEffect(() => {
    const fetchData = async () => {
      let data = await getQuota();
      if (
        data.payload.user_quota.cpu === 0 &&
        data.payload.user_quota.memory === 0
      ) {
        setShowModal(true);
        setQuotaMessage(
          "Quota is not set for the user. Please request access or wait until you are added to a group to use the Catalog."
        );
        setPrimaryButtonText("Go to Dashboard");
        setNavigateTo("/");
      } else if (
        data.payload.available_quota.cpu === 0 ||
        data.payload.available_quota.memory === 0
      ) {
        setShowModal(true);
        setQuotaMessage(
          "You are running out of quota, please make some space by deleting services or request to a higher quota group"
        );
        setPrimaryButtonText("Go to Dashboard");
        setNavigateTo("/");
      }
    };
    fetchData();
  }, []);

  const onCancel = async () => {
    setShowModal(false);
  };

  return (
    <Modal
      size="sm"
      onRequestClose={onCancel}
      onRequestSubmit={() => {
        navigate(navigateTo);
      }}
      open={showModal}
      modalHeading={quotaMessage}
      modalLabel="Quota Warning"
      primaryButtonText={primaryButtonText}
      secondaryButtonText={"Cancel"}
    ></Modal>
  );
};

export default QuotaWarning;
