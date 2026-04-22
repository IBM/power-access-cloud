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
    // Extract username and email from the original row data
    const rowData = selectRows[0];
    
    // Merge row data with API response - API response first, then override with row data
    const mergedData = {
      ...(result?.payload || result),
      username: rowData?.username,
      email: rowData?.email
    };
    setRequest(mergedData);
  };

  const renderValue = (value, level = 0) => {
    if (value === null || value === undefined) {
      return <span style={{ color: '#333' }}>null</span>;
    }
    
    if (Array.isArray(value)) {
      if (value.length === 0) {
        return <span style={{ color: '#333' }}>[]</span>;
      }
      return (
        <div>
          {value.map((item, index) => (
            <div key={index} style={{ marginTop: '4px' }}>
              <span style={{ color: '#333' }}>[{index}]:</span> {renderValue(item, level)}
            </div>
          ))}
        </div>
      );
    }
    
    if (typeof value === 'object') {
      return renderObject(value, level);
    }
    
    return <span style={{ color: '#333' }}>{String(value)}</span>;
  };

  const renderObject = (obj, level = 0) => {
    if (!obj) return null;
    
    const indent = level * 20;
    
    return (
      <div>
        {Object.entries(obj).map(([key, value]) => (
          <div key={key} style={{ paddingLeft: `${indent}px`, marginTop: '8px' }}>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'flex-start' }}>
              <span style={{ fontWeight: '600', color: '#161616', minWidth: '120px' }}>
                {key}:
              </span>
              <div style={{ flex: 1 }}>
                {renderValue(value, level)}
              </div>
            </div>
          </div>
        ))}
      </div>
    );
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
        <div style={{ fontFamily: 'IBM Plex Mono, monospace', fontSize: '14px', lineHeight: '1.6' }}>
          {renderObject(request?.payload || request)}
        </div>
      )}
    </Modal>
  );
};

export default RequestDetails;
