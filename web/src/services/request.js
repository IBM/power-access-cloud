import UserService from "./UserService";
import axios from "axios";

const _axios = axios.create();
_axios.interceptors.request.use((config) => {
  if (UserService.isLoggedIn()) {
    const cb = () => {
      config.headers.Authorization = `Bearer ${UserService.getToken()}`;
      return Promise.resolve(config);
    };
    return UserService.updateToken(cb);
  }
});

export const allGroups = () => {
  const url = "/pac-go-server/groups";

  return _axios.get(url)
    .then(response => ({
      type: "LIST_GROUPS",
      payload: response.data
    }))
    .catch(error => ({
      type: "API_ERROR",
      payload: error
    }));
};

export const getGroup = (id)=>{
  const url = `/pac-go-server/groups/${id}`;

  return _axios.get(url)
  .then(response => ({
    type: "LIST_GROUP",
    payload: response.data
  }))
  .catch(error => ({
    type: "API_ERROR",
    payload: error
  }));
};

export const newRequest = (group) => {
  const url = `/pac-go-server/groups/${group.id}/request`;

  const requestData = {
    justification: group.justification,
  };
  return _axios.post(url, requestData)
    .then((response) => ({
      type: "LIST_GROUP",
      payload: response.data,
    }))
    .catch((error) => ({
      type: "API_ERROR",
      payload: error,
    }));
};

export const deleteGroup = (group) => {
  const url = `/pac-go-server/groups/${group.id}`;

  return _axios.post(url, group)
    .then((response) => ({
      type: "DELETE_GROUP",
      payload: response.data,
    }))
    .catch((error) => ({
      type: "API_ERROR",
      payload: error,
    }));
};

export const allRequests = () => {
  const url = "/pac-go-server/requests";

  return _axios.get(url)
    .then(response => ({
      type: "LIST_REQUESTS",
      payload: response.data
    }))
    .catch(error => ({
      type: "API_ERROR",
      payload: error
    }));
};

export const approveRequest = (id) => {
  const url = `/pac-go-server/requests/${id}/approve`;

  return _axios.post(url,null)
    .then(response => ({
      type: "APPROVE_REQUEST",
      payload: response.data
    }))
    .catch(error => ({
      type: "API_ERROR",
      payload: error
    }));
};

