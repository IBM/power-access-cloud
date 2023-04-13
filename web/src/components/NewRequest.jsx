// import axios from "axios";
import React, { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {useSelector, useDispatch} from 'react-redux'
import { newRequest } from "../modules/group";
import { getGroup } from "../modules/group";

export default function NewRequest() {

  const dispatch = useDispatch();
  const { group } = useSelector((state) => state);
  const { id } = useParams();

  useEffect(() => {
    loadUser();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps


  let navigate = useNavigate();

  const [g, setGroup] = useState({
    name: "",
    id: "",
    justification: "",
  });

  const loadUser = async () => {
    const result = dispatch(getGroup(id));;
    setGroup(result.data);
  };

  const onInputChange = (e) => {
    setGroup({ ...group, [e.target.name]: e.target.value });
  };

  const onSubmit = async (e) => {
    e.preventDefault();
    dispatch(newRequest(g));
    navigate("..");
  };

  return (
    <div className="container">
      <div className="row">
        <div className="col-md-6 offset-md-3 border rounded p-4 mt-2 shadow">
          <h2 className="text-center m-4">Request for a group</h2>

          <form onSubmit={(e) => onSubmit(e)}>
          <div className="mb-3">
              <label htmlFor="ID" className="form-label">
                ID
              </label>
              <input
                type={"text"}
                className="form-control"
                placeholder="Enter your name"
                name="id"
                value={group.id}
                // onChange={(e) => onInputChange(e)}
                disabled
              />
            </div>
            <div className="mb-3">
              <label htmlFor="Name" className="form-label">
                Name
              </label>
              <input
                type={"text"}
                className="form-control"
                placeholder="Enter your name"
                name="name"
                value={group.name}
                disabled
              />
            </div>
            <div className="mb-3">
              <label htmlFor="Justifcation" className="form-label">
              Justifcation
              </label>
              <textarea
                type={"text"}
                className="form-control"
                placeholder="Enter your Justifcation for joining the group"
                name="justification"
                value={group.justification}
                onChange={(e) => onInputChange(e)}
              />
            </div>
            <button type="submit" className="btn btn-outline-primary">
              Submit
            </button>
            <Link className="btn btn-outline-danger mx-2" to="/groups">
              Cancel
            </Link>
          </form>
        </div>
      </div>
    </div>
  );
}
