import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { allGroups, deleteGroup } from "../modules/groups";
// import { addRequest } from "../modules/requests";
import { Link } from "react-router-dom";


const GroupList = () => {
  const dispatch = useDispatch();
  const { groups } = useSelector((state) => state);

  useEffect(() => {
    dispatch(allGroups());
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="row">
      <div className="col-sm-12">
        <h1>Groups:</h1>
        <table className="table table-striped">
          <thead>
            <tr>
              <th>ID</th>
              <th>Group</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {groups.map((group) => (
              <tr key={group.id}>
                <td>{group.id}</td>
                <td>{group.name}</td>
                <td>
                  {/* <button
                    style={{ marginRight: "10px" }}
                    className="btn btn-xs btn-primary"
                    onClick={() => dispatch(addRequest(group))}
                  >
                    Request Group
                  </button> */}
                                    <Link
                    className="btn btn-outline-primary mx-2"
                    to={`/request/${group.id}`}
                  >
                    Request
                  </Link>
                  <button
                    className="btn btn-xs btn-danger"
                    onClick={() => dispatch(deleteGroup(group))}
                  >
                    Exit Group
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default GroupList;
