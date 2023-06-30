import React, { useState, useEffect } from "react";
import { Modal } from "@carbon/react";
import { useNavigate } from "react-router-dom";
import { getQuota } from "../../services/request";

const QuotaWarning = () => {
  const [showModal, setShowModal] = useState(false);
  const [quotaMessage, setQuotaMessage] = useState("");
  const [primaryButtonText, setPrimaryButtonText] = useState("");
  const [nagivateTo, setNagivateTo] = useState("");
  const [passiveModal, setPassiveModal] = useState(false);
  let navigate = useNavigate();
  const onSubmit = async () => {
    navigate("/groups");
  };

  useEffect(() => {
    const fetchData = async () => {
      let data = await getQuota();
      if (
        data.payload.user_quota.cpu === 0 &&
        data.payload.user_quota.memory === 0
      ) {
        setShowModal(true);
        setQuotaMessage(
          "Quota is not set for the user, please request to any group to use catalogs"
        );
        setPrimaryButtonText("Request Group");
        setNagivateTo("/groups");
      } else if (
        data.payload.available_quota.cpu === 0 ||
        data.payload.available_quota.memory === 0
      ) {
        setShowModal(true);
        setQuotaMessage(
          "You are running out of quota, please make some space by deleting services or request to a higher quota group"
        );
        setPrimaryButtonText("Delete Services");
        setNagivateTo("/services");
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
      passiveModal={passiveModal}
      onRequestClose={onCancel}
      onRequestSubmit={() => {
        navigate(nagivateTo);
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
