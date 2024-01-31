import { SUCCESS_SUFFIX } from "redux-axios-middleware";
import UserService from "../services/UserService";
import { APPROVE_REQUEST,
        REJECT_REQUEST, 
        LIST_REQUESTS, 
        ADD_REQUEST, 
        LIST_GROUP, 
        NEW_REQUEST, 
        LIST_GROUPS, 
        DELETE_GROUP } from "./actionConstants";


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
