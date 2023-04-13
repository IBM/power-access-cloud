import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { allRequests } from "../modules/requests";

const RequestList = () => {

  const dispatch = useDispatch();
  const { requests } = useSelector((state) => state);

  useEffect(() => {
    dispatch(allRequests());
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="row">
      <div className="col-sm-12">
        <h1>Requests:</h1>
        <table className="table table-striped">
          <thead>
            <tr>
              <th>ID</th>
              <th>Group</th>
              <th>Requester</th>
              <th>Created At</th>
            </tr>
          </thead>
          <tbody>
            {requests.map((request) => {
              return (
                <tr key={request.id}>
                  <td>{request.id}</td>
                  <td>{request.group}</td>
                  <td>{request.requester}</td>
                  <td>{request.created_at}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default RequestList;
