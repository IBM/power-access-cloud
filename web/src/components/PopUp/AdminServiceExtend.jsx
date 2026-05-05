import React, { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { DatePicker, DatePickerInput, Modal, TextArea } from "@carbon/react";
import { extendServices } from "../../services/request";

const AdminServiceExtend = ({ pagename, selectRows, setActionProps, response }) => {
  const navigate = useNavigate();
  const service = selectRows?.[0];
  const serviceName = service?.name;

  const initialExpiryDate = useMemo(() => {
    if (!service?.expiry) {
      return null;
    }

    const parsedDate = new Date(service.expiry);
    return Number.isNaN(parsedDate.getTime()) ? null : parsedDate;
  }, [service]);

  const [selectedDate, setSelectedDate] = useState(initialExpiryDate);
  const [justification, setJustification] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const closeModal = () => {
    if (!submitting) {
      setActionProps("");
    }
  };

  const onSubmit = async () => {
    let title = "";
    let message = "";
    let errored = false;

    if (!serviceName) {
      response("Service extension failed", "Service details are unavailable.", true);
      setActionProps("");
      return;
    }

    if (!justification.trim()) {
      response("Service extension failed", "Please enter a justification.", true);
      return;
    }

    if (!selectedDate || Number.isNaN(selectedDate.getTime())) {
      response("Service extension failed", "Please select a valid expiry date.", true);
      return;
    }

    try {
      setSubmitting(true);

      const { type, payload } = await extendServices(serviceName, {
        justification: justification.trim(),
        type: "SERVICE_EXPIRY",
        service: {
          name: serviceName,
          expiry: selectedDate.toISOString(),
        },
      });

      if (type === "API_ERROR") {
        title = "Service extension failed";
        message = payload.response?.data?.error || "An error occurred";
        errored = true;
      } else {
        title =
          "Service extension request submitted successfully, please wait for the approval. For more details please check status section under My Services Tab";
      }
    } catch (error) {
      title = "Service extension failed";
      message = error.message || "An unexpected error occurred";
      errored = true;
    }

    setSubmitting(false);
    response(title, message, errored);
    setActionProps("");

    if (!errored && pagename) {
      navigate(pagename);
    }
  };

  return (
    <Modal
      open={true}
      modalLabel="Change Service Expiry"
      modalHeading={`Change expiry for "${service?.display_name || serviceName || "service"}"`}
      onRequestClose={closeModal}
      onRequestSubmit={onSubmit}
      primaryButtonText={submitting ? "Submitting..." : "Submit"}
      secondaryButtonText="Cancel"
      primaryButtonDisabled={submitting}
    >
      <div>
        <p style={{ marginBottom: "1rem" }}>
          Choose a new expiry date for this service and submit the request.
        </p>

        <div className="mb-3">
          <DatePicker
            datePickerType="single"
            dateFormat="m/d/Y"
            value={selectedDate || undefined}
            minDate={new Date()}
            onChange={(dates) => {
              setSelectedDate(dates?.[0] || null);
            }}
          >
            <DatePickerInput
              id="admin-service-expiry-date"
              labelText="New expiry date"
              placeholder="mm/dd/yyyy"
            />
          </DatePicker>
        </div>

        <div className="mb-3" style={{ marginTop: "1rem" }}>
          <TextArea
            id="admin-service-expiry-justification"
            labelText="Justification"
            placeholder="Enter your justification for changing the service expiry."
            value={justification}
            onChange={(e) => setJustification(e.target.value)}
          />
        </div>
      </div>
    </Modal>
  );
};

export default AdminServiceExtend;
