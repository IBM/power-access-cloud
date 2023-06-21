import { SUCCESS_SUFFIX } from "redux-axios-middleware";
import UserService from "../services/UserService";

const APPROVE_REQUEST = "APPROVE_REQUEST";
const REJECT_REQUEST = "REJECT_REQUEST";
const USER_LOGIN = "USER_LOGIN";
const LIST_GROUP = "LIST_GROUP";
const NEW_REQUEST = "NEW_REQUEST";
const LIST_GROUPS = "LIST_GROUPS";
const DELETE_GROUP = "DELETE_GROUP";
const LIST_REQUESTS = "LIST_REQUESTS";
const ADD_REQUEST = "ADD_REQUEST"



export const approveRequest = (id) => ({
  type: APPROVE_REQUEST,
  payload: {
    request: {
      url: `/pac-go-server/request/${id}/approve`,
      method: "POST",
    },
  },
});

export const rejectRequest = (id) => ({
  type: REJECT_REQUEST,
  payload: {
    request: {
      url: `/pac-go-server/request/${id}/reject`,
      method: "POST",
    },
  },
});


export const allRequests = () => ({
  type: LIST_REQUESTS,
  payload: {
    request: {
      url: "/pac-go-server/requests",
    },
  },
});

export const addRequest = (group) => {
  const username = UserService.getUsername();
  console.log(`${username} added the request ${group.name}`);

  const requestData = {
    group: group.name,
    requester: username,
  };

  return {
    type: ADD_REQUEST,
    payload: {
      request: {
        url: "/pac-go-server/request",
        method: "POST",
        data: requestData,
      },
    },
  };
};


export const getGroup = (id) => ({
  type: LIST_GROUP,
  payload: {
    request: {
      url: `/pac-go-server/group/${id}`,
    },
  },
});

export const newRequest = (group) => {
    const username = UserService.getUsername();
    console.log(`${username} added the request ${group.name}`);
  
    const requestData = {
      justification: group.justification,
    };
  
    return {
      type: NEW_REQUEST,
      payload: {
        request: {
          url: `/pac-go-server/group/${group.id}/request`,
          method: "POST",
          data: requestData,
        },
      },
    };
  };


  export const allGroups = () => ({
    type: LIST_GROUPS,
    payload: {
      request: {
        url: "/pac-go-server/groups",
      },
    },
  });
  
  export const deleteGroup = (group) => {
    console.log(`${UserService.getUsername()} exits the group ${group.name}`);
    return {
      type: DELETE_GROUP,
      payload: {
        group,
        request: {
          url: `/pac-go-server/groups/${group.id}`,
          method: "DELETE",
        },
      },
    };
  };
