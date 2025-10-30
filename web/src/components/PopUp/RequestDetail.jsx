// import axios from "axios";
import React, { useEffect, useState } from "react";
import { getRequest } from "../../services/request";
import { Modal } from "@carbon/react";

const RequestDetails = ({ selectRows, setActionProps }) => {
  const [loading, setLoading] = useState(true);
  const id = selectRows[0]?.id;
  useEffect(() => {
    loadRequest();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const [request, setRequest] = useState();

  const loadRequest = async () => {
    const result = await getRequest(id);
    setLoading(false);
    setRequest(result);
  };

  return (
    <Modal
      modalHeading="Request Details"
      onRequestClose={() => {
        setActionProps("");
      }}
      open={true}
      passiveModal={true}
    >
      {loading && <>Loading....</>}
      {!loading && (
        <div>
          <div className="mb-3">
            <div>
              <pre>{JSON.stringify(request, undefined, 2)}</pre>
            </div>
          </div>
        </div>
      )}
    </Modal>
  );
};

export default RequestDetails;
